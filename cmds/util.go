package cmds

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"
	"github.com/ProtoconNet/mitum-currency/v3/operation/processor"
	quicstreamheader "github.com/ProtoconNet/mitum2/network/quicstream/header"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	mongodbstorage "github.com/ProtoconNet/mitum-currency/v3/digest/mongodb"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	isaacblock "github.com/ProtoconNet/mitum2/isaac/block"
	isaacdatabase "github.com/ProtoconNet/mitum2/isaac/database"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	isaacoperation "github.com/ProtoconNet/mitum2/isaac/operation"
	isaacstates "github.com/ProtoconNet/mitum2/isaac/states"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/network/quicstream"
	"github.com/ProtoconNet/mitum2/storage"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/ProtoconNet/mitum2/util/ps"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	PNameDigestDesign                   = ps.Name("digest-design")
	PNameOperationProcessorsMap         = ps.Name("mitum-currency-operation-processors-map")
	PNameGenerateGenesis                = ps.Name("mitum-currency-generate-genesis")
	PNameDigestAPIHandlers              = ps.Name("mitum-currency-digest-api-handlers")
	PNameDigesterFollowUp               = ps.Name("mitum-currency-followup_digester")
	BEncoderContextKey                  = util.ContextKey("bson-encoder")
	ProposalOperationFactHintContextKey = util.ContextKey("proposal-operation-fact--hint")
)

type ProposalOperationFactHintFunc func() func(hint.Hint) bool

func LoadFromStdInput() ([]byte, error) {
	var b []byte
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			b = append(b, sc.Bytes()...)
			b = append(b, []byte("\n")...)
		}

		if err := sc.Err(); err != nil {
			return nil, err
		}
	}

	return bytes.TrimSpace(b), nil
}

func GenerateED25519PrivateKey() (ed25519.PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)

	return priv, err
}

func GenerateTLSCertsPair(host string, key ed25519.PrivateKey) (*pem.Block, *pem.Block, error) {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		DNSNames:     []string{host},
		NotBefore:    time.Now().Add(time.Minute * -1),
		NotAfter:     time.Now().Add(time.Hour * 24 * 1825),
	}

	if i := net.ParseIP(host); i != nil {
		template.IPAddresses = []net.IP{i}
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		key.Public().(ed25519.PublicKey),
		key,
	)
	if err != nil {
		return nil, nil, err
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}

	return &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes},
		&pem.Block{Type: "CERTIFICATE", Bytes: certDER},
		nil
}

func GenerateTLSCerts(host string, key ed25519.PrivateKey) ([]tls.Certificate, error) {
	k, c, err := GenerateTLSCertsPair(host, key)
	if err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(pem.EncodeToMemory(c), pem.EncodeToMemory(k))
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{certificate}, nil
}

type NetworkIDFlag []byte

func (v *NetworkIDFlag) UnmarshalText(b []byte) error {
	*v = b

	return nil
}

func (v NetworkIDFlag) NetworkID() base.NetworkID {
	return base.NetworkID(v)
}

func PrettyPrint(out io.Writer, i interface{}) {
	var b []byte
	b, err := enc.Marshal(i)
	if err != nil {
		panic(err)
	}

	_, _ = fmt.Fprintln(out, string(b))
}

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return pctx, err
	}

	limiterF, err := launch.NewSuffrageCandidateLimiterFunc(pctx)
	if err != nil {
		return pctx, err
	}

	set := hint.NewCompatibleSet()

	opr := processor.NewOperationProcessor()
	if err := opr.SetProcessor(
		currency.CreateAccountsHint,
		currency.NewCreateAccountsProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.KeyUpdaterHint,
		currency.NewKeyUpdaterProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.TransfersHint,
		currency.NewTransfersProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.CurrencyRegisterHint,
		currency.NewCurrencyRegisterProcessor(isaacParams.Threshold()),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.CurrencyPolicyUpdaterHint,
		currency.NewCurrencyPolicyUpdaterProcessor(isaacParams.Threshold()),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.SuffrageInflationHint,
		currency.NewSuffrageInflationProcessor(isaacParams.Threshold()),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		extension.CreateContractAccountsHint,
		extension.NewCreateContractAccountsProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		extension.WithdrawsHint,
		extension.NewWithdrawsProcessor(),
	); err != nil {
		return pctx, err
	}

	_ = set.Add(currency.CreateAccountsHint, func(height base.Height) (base.OperationProcessor, error) {
		return opr.New(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(currency.KeyUpdaterHint, func(height base.Height) (base.OperationProcessor, error) {
		return opr.New(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(currency.TransfersHint, func(height base.Height) (base.OperationProcessor, error) {
		return opr.New(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(currency.CurrencyRegisterHint, func(height base.Height) (base.OperationProcessor, error) {
		return opr.New(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(currency.CurrencyPolicyUpdaterHint, func(height base.Height) (base.OperationProcessor, error) {
		return opr.New(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(currency.SuffrageInflationHint, func(height base.Height) (base.OperationProcessor, error) {
		return opr.New(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(extension.CreateContractAccountsHint, func(height base.Height) (base.OperationProcessor, error) {
		return opr.New(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(extension.WithdrawsHint, func(height base.Height) (base.OperationProcessor, error) {
		return opr.New(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(isaacoperation.SuffrageCandidateHint, func(height base.Height) (base.OperationProcessor, error) {
		policy := db.LastNetworkPolicy()
		if policy == nil { // NOTE Usually it means empty block data
			return nil, nil
		}

		return isaacoperation.NewSuffrageCandidateProcessor(
			height,
			db.State,
			limiterF,
			nil,
			policy.SuffrageCandidateLifespan(),
		)
	})

	_ = set.Add(isaacoperation.SuffrageJoinHint, func(height base.Height) (base.OperationProcessor, error) {
		policy := db.LastNetworkPolicy()
		if policy == nil { // NOTE Usually it means empty block data
			return nil, nil
		}

		return isaacoperation.NewSuffrageJoinProcessor(
			height,
			isaacParams.Threshold(),
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(isaac.SuffrageExpelOperationHint, func(height base.Height) (base.OperationProcessor, error) {
		policy := db.LastNetworkPolicy()
		if policy == nil { // NOTE Usually it means empty block data
			return nil, nil
		}

		return isaacoperation.NewSuffrageExpelProcessor(
			height,
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(isaacoperation.SuffrageDisjoinHint, func(height base.Height) (base.OperationProcessor, error) {
		return isaacoperation.NewSuffrageDisjoinProcessor(
			height,
			db.State,
			nil,
			nil,
		)
	})

	var f ProposalOperationFactHintFunc = IsSupportedProposalOperationFactHintFunc

	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, set) //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, ProposalOperationFactHintContextKey, f)

	return pctx, nil
}

func PGenerateGenesis(pctx context.Context) (context.Context, error) {
	e := util.StringError("failed to generate genesis block")

	var log *logging.Logging
	var design launch.NodeDesign
	var genesisDesign launch.GenesisDesign
	var enc encoder.Encoder
	var local base.LocalNode
	var isaacParams *isaac.Params
	var db isaac.Database

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.DesignContextKey, &design,
		launch.GenesisDesignContextKey, &genesisDesign,
		launch.EncoderContextKey, &enc,
		launch.LocalContextKey, &local,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return pctx, e.Wrap(err)
	}

	g := NewGenesisBlockGenerator(
		local,
		isaacParams.NetworkID(),
		enc,
		db,
		launch.LocalFSDataDirectory(design.Storage.Base),
		genesisDesign.Facts,
	)
	_ = g.SetLogging(log)

	if _, err := g.Generate(); err != nil {
		return pctx, e.Wrap(err)
	}

	return pctx, nil
}

func PEncoder(pctx context.Context) (context.Context, error) {
	e := util.StringError("failed to prepare encoders")

	encs := encoder.NewEncoders()
	jenc := jsonenc.NewEncoder()
	benc := bsonenc.NewEncoder()

	if err := encs.AddHinter(jenc); err != nil {
		return pctx, e.Wrap(err)
	}
	if err := encs.AddHinter(benc); err != nil {
		return pctx, e.Wrap(err)
	}

	pctx = context.WithValue(pctx, launch.EncodersContextKey, encs) //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, launch.EncoderContextKey, jenc)  //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, BEncoderContextKey, benc)        //revive:disable-line:modifies-parameter

	return pctx, nil
}

func PLoadDigestDesign(pctx context.Context) (context.Context, error) {
	e := util.StringError("failed to load design")

	var log *logging.Logging
	var flag launch.DesignFlag
	var enc *jsonenc.Encoder

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.DesignFlagContextKey, &flag,
		launch.EncoderContextKey, &enc,
	); err != nil {
		return pctx, e.Wrap(err)
	}

	switch flag.Scheme() {
	case "file":
		b, err := os.ReadFile(filepath.Clean(flag.URL().Path))
		if err != nil {
			return pctx, e.Wrap(err)
		}

		var m struct {
			Digest *DigestDesign
		}

		if err := yaml.Unmarshal(b, &m); err != nil {
			return pctx, err
		} else if m.Digest == nil {
			return pctx, nil
		} else if i, err := m.Digest.Set(pctx); err != nil {
			return pctx, err
		} else {
			pctx = i
		}

		pctx = context.WithValue(pctx, ContextValueDigestDesign, *m.Digest)

		log.Log().Debug().Object("design", *m.Digest).Msg("digest design loaded")
	default:
		return pctx, e.Errorf("unknown digest design uri, %q", flag.URL())
	}

	return pctx, nil
}

func PNetworkHandlers(pctx context.Context) (context.Context, error) {
	e := util.StringError("failed to prepare network handlers")

	var log *logging.Logging
	var encs *encoder.Encoders
	var enc encoder.Encoder
	var design launch.NodeDesign
	var local base.LocalNode
	var params *launch.LocalParams
	var db isaac.Database
	var pool *isaacdatabase.TempPool
	var proposalMaker *isaac.ProposalMaker
	var m *quicmemberlist.Memberlist
	var syncSourcePool *isaac.SyncSourcePool
	var handlers *quicstream.PrefixHandler
	var nodeInfo *isaacnetwork.NodeInfoUpdater
	var svVoteF isaac.SuffrageVoteFunc
	var ballotBox *isaacstates.Ballotbox
	var filterNotifyMsg quicmemberlist.FilterNotifyMsgFunc

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.EncodersContextKey, &encs,
		launch.EncoderContextKey, &enc,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
		launch.PoolDatabaseContextKey, &pool,
		launch.ProposalMakerContextKey, &proposalMaker,
		launch.MemberlistContextKey, &m,
		launch.SyncSourcePoolContextKey, &syncSourcePool,
		launch.QuicstreamHandlersContextKey, &handlers,
		launch.NodeInfoContextKey, &nodeInfo,
		launch.SuffrageVotingVoteFuncContextKey, &svVoteF,
		launch.BallotboxContextKey, &ballotBox,
		launch.FilterMemberlistNotifyMsgFuncContextKey, &filterNotifyMsg,
	); err != nil {
		return pctx, e.Wrap(err)
	}

	isaacParams := params.ISAAC

	lastBlockMapF := launch.QuicstreamHandlerLastBlockMapFunc(db)
	suffrageNodeConnInfoF := launch.QuicstreamHandlerSuffrageNodeConnInfoFunc(db, m)

	handlers.
		Add(isaacnetwork.HandlerPrefixLastSuffrageProof,
			quicstreamheader.NewHandler(encs, 0, isaacnetwork.QuicstreamHandlerLastSuffrageProof(
				func(last util.Hash) (hint.Hint, []byte, []byte, bool, error) {
					encHint, metaBytes, body, found, lastHeight, err := db.LastSuffrageProofBytes()

					switch {
					case err != nil:
						return encHint, nil, nil, false, err
					case !found:
						return encHint, nil, nil, false, storage.ErrNotFound.Errorf("last SuffrageProof not found")
					}

					switch h, err := isaacdatabase.ReadHashRecordMeta(metaBytes); {
					case err != nil:
						return encHint, nil, nil, true, err
					case last != nil && last.Equal(h):
						nBody, _ := util.NewLengthedBytesSlice(0x01, [][]byte{lastHeight.Bytes(), nil})

						return encHint, nil, nBody, false, nil
					default:
						nBody, _ := util.NewLengthedBytesSlice(0x01, [][]byte{lastHeight.Bytes(), body})

						return encHint, metaBytes, nBody, true, nil
					}
				},
			), nil)).
		Add(isaacnetwork.HandlerPrefixSuffrageProof,
			quicstreamheader.NewHandler(encs, 0,
				isaacnetwork.QuicstreamHandlerSuffrageProof(db.SuffrageProofBytes), nil)).
		Add(isaacnetwork.HandlerPrefixLastBlockMap,
			quicstreamheader.NewHandler(encs, 0, isaacnetwork.QuicstreamHandlerLastBlockMap(lastBlockMapF), nil)).
		Add(isaacnetwork.HandlerPrefixBlockMap,
			quicstreamheader.NewHandler(encs, 0, isaacnetwork.QuicstreamHandlerBlockMap(db.BlockMapBytes), nil)).
		Add(isaacnetwork.HandlerPrefixBlockMapItem,
			quicstreamheader.NewHandler(encs, 0,
				isaacnetwork.QuicstreamHandlerBlockMapItem(
					func(height base.Height, item base.BlockMapItemType) (io.ReadCloser, bool, error) {
						e := util.StringError("get BlockMapItem")

						var menc encoder.Encoder

						switch m, found, err := db.BlockMap(height); {
						case err != nil:
							return nil, false, e.Wrap(err)
						case !found:
							return nil, false, e.Wrap(storage.ErrNotFound.Errorf("BlockMap not found"))
						default:
							menc = encs.Find(m.Encoder())
							if menc == nil {
								return nil, false, e.Wrap(storage.ErrNotFound.Errorf("encoder of BlockMap not found"))
							}
						}

						reader, err := isaacblock.NewLocalFSReaderFromHeight(
							launch.LocalFSDataDirectory(design.Storage.Base), height, menc,
						)
						if err != nil {
							return nil, false, e.Wrap(err)
						}
						defer func() {
							_ = reader.Close()
						}()

						return reader.Reader(item)
					},
				), nil)).
		Add(isaacnetwork.HandlerPrefixNodeChallenge,
			quicstreamheader.NewHandler(encs, 0,
				isaacnetwork.QuicstreamHandlerNodeChallenge(isaacParams.NetworkID(), local), nil)).
		Add(isaacnetwork.HandlerPrefixSuffrageNodeConnInfo,
			quicstreamheader.NewHandler(encs, 0,
				isaacnetwork.QuicstreamHandlerSuffrageNodeConnInfo(suffrageNodeConnInfoF), nil)).
		Add(isaacnetwork.HandlerPrefixSyncSourceConnInfo,
			quicstreamheader.NewHandler(encs, 0, isaacnetwork.QuicstreamHandlerSyncSourceConnInfo(
				func() ([]isaac.NodeConnInfo, error) {
					members := make([]isaac.NodeConnInfo, syncSourcePool.Len()*2)

					var i int
					syncSourcePool.Actives(func(nci isaac.NodeConnInfo) bool {
						members[i] = nci
						i++

						return true
					})

					return members[:i], nil
				},
			), nil)).
		Add(isaacnetwork.HandlerPrefixState,
			quicstreamheader.NewHandler(encs, 0, isaacnetwork.QuicstreamHandlerState(db.StateBytes), nil)).
		Add(isaacnetwork.HandlerPrefixExistsInStateOperation,
			quicstreamheader.NewHandler(encs, 0,
				isaacnetwork.QuicstreamHandlerExistsInStateOperation(db.ExistsInStateOperation), nil)).
		Add(isaacnetwork.HandlerPrefixNodeInfo,
			quicstreamheader.NewHandler(encs, 0, isaacnetwork.QuicstreamHandlerNodeInfo(
				launch.QuicstreamHandlerGetNodeInfoFunc(enc, nodeInfo)), nil)).
		Add(isaacnetwork.HandlerPrefixSendBallots,
			quicstreamheader.NewHandler(encs, 0, isaacnetwork.QuicstreamHandlerSendBallots(
				isaacParams.NetworkID(),
				func(bl base.BallotSignFact) error {
					switch passed, err := filterNotifyMsg(bl); {
					case err != nil:
						log.Log().Trace().
							Str("module", "filter-notify-msg-send-ballots").
							Err(err).
							Interface("handover_message", bl).
							Msg("filter error")

						fallthrough
					case !passed:
						log.Log().Trace().
							Str("module", "filter-notify-msg-send-ballots").
							Interface("handover_message", bl).
							Msg("filtered")

						return nil
					}

					_, err := ballotBox.VoteSignFact(bl)

					return err
				},
				params.MISC.MaxMessageSize,
			), nil))

	if err := launch.AttachMemberlistNetworkHandlers(pctx); err != nil {
		return pctx, err
	}

	return pctx, nil
}

func PStatesNetworkHandlers(pctx context.Context) (context.Context, error) {
	var encs *encoder.Encoders
	var local base.LocalNode
	var isaacParams *isaac.Params
	var handlers *quicstream.PrefixHandler
	var states *isaacstates.States

	if err := util.LoadFromContext(pctx,
		launch.EncodersContextKey, &encs,
		launch.LocalContextKey, &local,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.QuicstreamHandlersContextKey, &handlers,
		launch.StatesContextKey, &states,
	); err != nil {
		return pctx, err
	}

	if err := launch.AttachHandlerOperation(pctx, handlers); err != nil {
		return pctx, err
	}

	if err := AttachHandlerSendOperation(pctx, handlers); err != nil {
		return pctx, err
	}

	if err := launch.AttachHandlerStreamOperations(pctx, handlers); err != nil {
		return pctx, err
	}

	if err := launch.AttachHandlerProposals(pctx, handlers); err != nil {
		return pctx, err
	}

	handlers.
		Add(isaacnetwork.HandlerPrefixSetAllowConsensus,
			quicstreamheader.NewHandler(encs,
				time.Second*2, //nolint:gomnd //...
				isaacnetwork.QuicstreamHandlerSetAllowConsensus(
					local.Publickey(),
					isaacParams.NetworkID(),
					states.SetAllowConsensus,
				),
				nil,
			),
		)

	return pctx, nil
}

func AttachHandlerSendOperation(pctx context.Context, handlers *quicstream.PrefixHandler) error {
	var log *logging.Logging
	var encs *encoder.Encoders
	var params *launch.LocalParams
	var db isaac.Database
	var pool *isaacdatabase.TempPool
	var states *isaacstates.States
	var svVoteF isaac.SuffrageVoteFunc
	var memberList *quicmemberlist.Memberlist

	if err := util.LoadFromContext(pctx,
		launch.LoggingContextKey, &log,
		launch.EncodersContextKey, &encs,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
		launch.PoolDatabaseContextKey, &pool,
		launch.StatesContextKey, &states,
		launch.SuffrageVotingVoteFuncContextKey, &svVoteF,
		launch.MemberlistContextKey, &memberList,
	); err != nil {
		return err
	}

	sendOperationFilterF, err := SendOperationFilterFunc(pctx)
	if err != nil {
		return err
	}

	handlers.Add(isaacnetwork.HandlerPrefixSendOperation, quicstreamheader.NewHandler(encs, 0,
		isaacnetwork.QuicstreamHandlerSendOperation(
			params.ISAAC.NetworkID(),
			pool,
			db.ExistsInStateOperation,
			sendOperationFilterF,
			svVoteF,
			func(ctx context.Context, id string, op base.Operation, b []byte) error {
				if broker := states.HandoverXBroker(); broker != nil {
					if err := broker.SendData(ctx, isaacstates.HandoverMessageDataTypeOperation, op); err != nil {
						log.Log().Error().Err(err).
							Interface("operation", op.Hash()).
							Msg("failed to send operation data to handover y broker; ignored")
					}
				}

				return memberList.CallbackBroadcast(b, id, nil)
			},
			params.MISC.MaxMessageSize,
		),
		nil))

	return nil
}

func SendOperationFilterFunc(ctx context.Context) (
	func(base.Operation) (bool, error),
	error,
) {
	var db isaac.Database
	var oprs *hint.CompatibleSet
	var f ProposalOperationFactHintFunc

	if err := util.LoadFromContextOK(ctx,
		launch.CenterDatabaseContextKey, &db,
		launch.OperationProcessorsMapContextKey, &oprs,
		ProposalOperationFactHintContextKey, &f,
	); err != nil {
		return nil, err
	}

	operationFilterF := f()

	return func(op base.Operation) (bool, error) {
		switch hinter, ok := op.Fact().(hint.Hinter); {
		case !ok:
			return false, nil
		case !operationFilterF(hinter.Hint()):
			return false, errors.Errorf("Not supported operation")
		}
		var height base.Height

		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return false, err
		case !found:
			return true, nil
		default:
			height = m.Manifest().Height()
		}

		f, closeF, err := launch.OperationPreProcess(oprs, op, height)
		if err != nil {
			return false, err
		}

		defer func() {
			_ = closeF()
		}()

		_, reason, err := f(context.Background(), db.State)
		if err != nil {
			return false, err
		}

		return reason == nil, reason
	}, nil
}

func IsSupportedProposalOperationFactHintFunc() func(hint.Hint) bool {
	return func(ht hint.Hint) bool {
		for i := range SupportedProposalOperationFactHinters {
			s := SupportedProposalOperationFactHinters[i].Hint
			if ht.Type() != s.Type() {
				continue
			}

			return ht.IsCompatible(s)
		}

		return false
	}
}

func ProcessDatabase(ctx context.Context) (context.Context, error) {
	var l DigestDesign
	if err := util.LoadFromContext(ctx, ContextValueDigestDesign, &l); err != nil {
		return ctx, err
	}

	if (l == DigestDesign{}) {
		return ctx, nil
	}
	conf := l.Database()

	switch {
	case conf.URI().Scheme == "mongodb", conf.URI().Scheme == "mongodb+srv":
		return processMongodbDatabase(ctx, l)
	default:
		return ctx, errors.Errorf("unsupported database type, %q", conf.URI().Scheme)
	}
}

func processMongodbDatabase(ctx context.Context, l DigestDesign) (context.Context, error) {
	conf := l.Database()

	/*
		ca, err := cache.NewCacheFromURI(conf.Cache().String())
		if err != nil {
			return ctx, err
		}
	*/

	var encs *encoder.Encoders
	if err := util.LoadFromContext(ctx, launch.EncodersContextKey, &encs); err != nil {
		return ctx, err
	}

	st, err := mongodbstorage.NewDatabaseFromURI(conf.URI().String(), encs)
	if err != nil {
		return ctx, err
	}

	if err := st.Initialize(); err != nil {
		return ctx, err
	}

	var db isaac.Database
	if err := util.LoadFromContextOK(ctx, launch.CenterDatabaseContextKey, &db); err != nil {
		return ctx, err
	}

	mst, ok := db.(*isaacdatabase.Center)
	if !ok {
		return ctx, errors.Errorf("expected isaacdatabase.Center, not %T", db)
	}

	dst, err := loadDigestDatabase(mst, st, false)
	if err != nil {
		return ctx, err
	}
	var log *logging.Logging
	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, err
	}

	_ = dst.SetLogging(log)

	return context.WithValue(ctx, ContextValueDigestDatabase, dst), nil
}
