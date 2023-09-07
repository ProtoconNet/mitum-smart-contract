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
	"io"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	isaacdatabase "github.com/ProtoconNet/mitum2/isaac/database"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	isaacstates "github.com/ProtoconNet/mitum2/isaac/states"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/ProtoconNet/mitum2/util/ps"
	"github.com/pkg/errors"
)

var (
	PNameDigestDesign                   = ps.Name("digest-design")
	PNameGenerateGenesis                = ps.Name("mitum-currency-generate-genesis")
	PNameDigestAPIHandlers              = ps.Name("mitum-currency-digest-api-handlers")
	PNameDigesterFollowUp               = ps.Name("mitum-currency-followup_digester")
	BEncoderContextKey                  = util.ContextKey("bson-encoder")
	ProposalOperationFactHintContextKey = util.ContextKey("proposal-operation-fact--hint")
	OperationProcessorContextKey        = util.ContextKey("mitum-currency-operation-processor")
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

func AttachHandlerSendOperation(pctx context.Context) error {
	var log *logging.Logging
	var params *launch.LocalParams
	var db isaac.Database
	var pool *isaacdatabase.TempPool
	var states *isaacstates.States
	var svVoteF isaac.SuffrageVoteFunc
	var memberList *quicmemberlist.Memberlist

	if err := util.LoadFromContext(pctx,
		launch.LoggingContextKey, &log,
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

	var gerror error

	launch.EnsureHandlerAdd(pctx, &gerror,
		isaacnetwork.HandlerPrefixSendOperationString,
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
							Msg("send operation data to handover y broker; ignored")
					}
				}

				return memberList.CallbackBroadcast(b, id, nil)
			},
			params.MISC.MaxMessageSize,
		),
		nil,
	)

	return gerror
}

func SendOperationFilterFunc(ctx context.Context) (
	func(base.Operation) (bool, error),
	error,
) {
	var db isaac.Database
	var oprs *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]
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
