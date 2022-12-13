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

	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	mongodbstorage "github.com/spikeekips/mitum-currency/digest/mongodb"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/isaac"
	isaacblock "github.com/spikeekips/mitum/isaac/block"
	isaacdatabase "github.com/spikeekips/mitum/isaac/database"
	isaacnetwork "github.com/spikeekips/mitum/isaac/network"
	isaacoperation "github.com/spikeekips/mitum/isaac/operation"
	isaacstates "github.com/spikeekips/mitum/isaac/states"
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/network/quicmemberlist"
	"github.com/spikeekips/mitum/network/quicstream"
	"github.com/spikeekips/mitum/storage"
	leveldbstorage "github.com/spikeekips/mitum/storage/leveldb"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/localtime"
	"github.com/spikeekips/mitum/util/logging"
	"github.com/spikeekips/mitum/util/ps"
)

var (
	PNameOperationProcessorsMap = ps.Name("mitum-currency-operation-processors-map")
	PNameGenerateGenesis        = ps.Name("mitum-currency-generate-genesis")
	PNameDigestAPIHandlers      = ps.Name("mitum-currency-digest-api-handlers")
	PNameDigesterFollowUp       = ps.Name("mitum-currency-followup_digester")
	BEncoderContextKey          = util.ContextKey("bencoder")
)

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

func GenerateED25519Privatekey() (ed25519.PrivateKey, error) {
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

func POperationProcessorsMap(ctx context.Context) (context.Context, error) {

	var params *isaac.LocalParams
	var db isaac.Database

	if err := util.LoadFromContextOK(ctx,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return ctx, err
	}

	limiterf, err := launch.NewSuffrageCandidateLimiterFunc(ctx)
	if err != nil {
		return ctx, err
	}

	set := hint.NewCompatibleSet()

	opr := currency.NewOperationProcessor()
	opr.SetProcessor(currency.CreateAccountsHint, currency.NewCreateAccountsProcessor())
	opr.SetProcessor(currency.KeyUpdaterHint, currency.NewKeyUpdaterProcessor())
	opr.SetProcessor(currency.TransfersHint, currency.NewTransfersProcessor())
	opr.SetProcessor(currency.CurrencyRegisterHint, currency.NewCurrencyRegisterProcessor(params.Threshold()))

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

	_ = set.Add(isaacoperation.SuffrageCandidateHint, func(height base.Height) (base.OperationProcessor, error) {
		policy := db.LastNetworkPolicy()
		if policy == nil { // NOTE Usually it means empty block data
			return nil, nil
		}

		return isaacoperation.NewSuffrageCandidateProcessor(
			height,
			db.State,
			limiterf,
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
			params.Threshold(),
			db.State,
			nil,
			nil,
		)
	})

	_ = set.Add(isaac.SuffrageWithdrawOperationHint, func(height base.Height) (base.OperationProcessor, error) {
		policy := db.LastNetworkPolicy()
		if policy == nil { // NOTE Usually it means empty block data
			return nil, nil
		}

		return isaacoperation.NewSuffrageWithdrawProcessor(
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

	ctx = context.WithValue(ctx, launch.OperationProcessorsMapContextKey, set) //revive:disable-line:modifies-parameter

	return ctx, nil
}

func PGenerateGenesis(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to generate genesis block")

	var log *logging.Logging
	var design NodeDesign
	var genesisDesign launch.GenesisDesign
	var enc encoder.Encoder
	var local base.LocalNode
	var params *isaac.LocalParams
	var db isaac.Database

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.DesignContextKey, &design,
		launch.GenesisDesignContextKey, &genesisDesign,
		launch.EncoderContextKey, &enc,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return ctx, e(err, "")
	}

	g := NewGenesisBlockGenerator(
		local,
		params.NetworkID(),
		enc,
		db,
		launch.LocalFSDataDirectory(design.Storage.Base),
		genesisDesign.Facts,
	)
	_ = g.SetLogging(log)

	if _, err := g.Generate(); err != nil {
		return ctx, e(err, "")
	}

	return ctx, nil
}

func PEncoder(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to prepare encoders")

	encs := encoder.NewEncoders()
	jenc := jsonenc.NewEncoder()
	benc := bsonenc.NewEncoder()

	if err := encs.AddHinter(jenc); err != nil {
		return ctx, e(err, "")
	}
	if err := encs.AddHinter(benc); err != nil {
		return ctx, e(err, "")
	}

	ctx = context.WithValue(ctx, launch.EncodersContextKey, encs) //revive:disable-line:modifies-parameter
	ctx = context.WithValue(ctx, launch.EncoderContextKey, jenc)  //revive:disable-line:modifies-parameter
	ctx = context.WithValue(ctx, BEncoderContextKey, benc)        //revive:disable-line:modifies-parameter

	return ctx, nil
}

func PLoadDesign(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to load design")

	var log *logging.Logging
	var flag launch.DesignFlag
	var enc *jsonenc.Encoder
	var privfromvault string

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.DesignFlagContextKey, &flag,
		launch.EncoderContextKey, &enc,
		launch.VaultContextKey, &privfromvault,
	); err != nil {
		return ctx, e(err, "")
	}

	var design NodeDesign
	var digestDesign DigestDesign

	switch flag.Scheme() {
	case "file":
		switch d, _, err := NodeDesignFromFile(flag.URL().Path, enc); {
		case err != nil:
			return ctx, e(err, "")
		default:
			design = d
		}

		if (design.Digest != DigestDesign{}) {
			if i, err := design.Digest.Set(ctx); err != nil {
				return ctx, err
			} else {
				ctx = i
			}

			digestDesign = design.Digest
			ctx = context.WithValue(ctx, ContextValueDigestDesign, digestDesign)
		}

		// switch di, _, err := DigestDesignFromFile(flag.URL().Path, enc); {
		// case err != nil:
		// 	return ctx, e(err, "")
		// default:
		// 	digestDesign = d.DigestDesign
		// }
	case "http", "https":
		switch d, err := NodeDesignFromHTTP(flag.URL().String(), flag.Properties().HTTPSTLSInsecure, enc); {
		case err != nil:
			return ctx, e(err, "")
		default:
			design = d
		}
	case "consul":
		switch d, err := NodeDesignFromConsul(flag.URL().Host, flag.URL().Path, enc); {
		case err != nil:
			return ctx, e(err, "")
		default:
			design = d
		}
	default:
		return ctx, e(nil, "unknown design uri, %q", flag.URL())
	}

	log.Log().Debug().Object("design", design).Msg("design loaded")

	if len(privfromvault) > 0 {
		priv, err := loadPrivatekeyFromVault(privfromvault, enc)
		if err != nil {
			return ctx, e(err, "")
		}

		log.Log().Debug().Interface("privatekey", priv.Publickey()).Msg("privatekey loaded from vault")

		design.Privatekey = priv
	}

	ctx = context.WithValue(ctx, launch.DesignContextKey, design) //revive:disable-line:modifies-parameter//revive:disable-line:modifies-parameter

	return ctx, nil
}

func PCheckDesign(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to check design")

	var log *logging.Logging
	var flag launch.DesignFlag
	var devflags launch.DevFlags
	var design NodeDesign

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.DesignFlagContextKey, &flag,
		launch.DevFlagsContextKey, &devflags,
		launch.DesignContextKey, &design,
	); err != nil {
		return ctx, e(err, "")
	}

	if err := design.IsValid(nil); err != nil {
		return ctx, e(err, "")
	}

	if err := design.Check(devflags); err != nil {
		return ctx, e(err, "")
	}

	log.Log().Debug().Object("design", design).Msg("design checked")

	//revive:disable:modifies-parameter
	ctx = context.WithValue(ctx, launch.DesignContextKey, design)
	ctx = context.WithValue(ctx, launch.LocalParamsContextKey, design.LocalParams)
	//revive:enable:modifies-parameter

	if err := launch.UpdateFromConsulAfterCheckDesign(ctx, flag); err != nil {
		return ctx, e(err, "")
	}

	return ctx, nil
}

func PLocal(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to load local")

	var log *logging.Logging
	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, e(err, "")
	}

	var design NodeDesign
	if err := util.LoadFromContextOK(ctx, launch.DesignContextKey, &design); err != nil {
		return ctx, e(err, "")
	}

	local, err := LocalFromDesign(design)
	if err != nil {
		return ctx, e(err, "")
	}

	log.Log().Info().Interface("local", local).Msg("local loaded")

	ctx = context.WithValue(ctx, launch.LocalContextKey, local) //revive:disable-line:modifies-parameter

	return ctx, nil
}

func LocalFromDesign(design NodeDesign) (base.LocalNode, error) {
	local := isaac.NewLocalNode(design.Privatekey, design.Address)

	if err := local.IsValid(nil); err != nil {
		return nil, err
	}

	return local, nil
}

func PCheckLocalFS(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to check localfs")

	var version util.Version
	var design NodeDesign
	var params *isaac.LocalParams
	var encs *encoder.Encoders
	var enc encoder.Encoder

	if err := util.LoadFromContextOK(ctx,
		launch.VersionContextKey, &version,
		launch.DesignContextKey, &design,
		launch.EncodersContextKey, &encs,
		launch.EncoderContextKey, &enc,
		launch.LocalParamsContextKey, &params,
	); err != nil {
		return ctx, e(err, "")
	}

	fsnodeinfo, err := launch.CheckLocalFS(params.NetworkID(), design.Storage.Base, enc)

	switch {
	case err == nil:
		if err = isaacblock.CleanBlockTempDirectory(launch.LocalFSDataDirectory(design.Storage.Base)); err != nil {
			return ctx, e(err, "")
		}
	case errors.Is(err, os.ErrNotExist):
		if err = launch.CleanStorage(
			design.Storage.Database.String(),
			design.Storage.Base,
			encs,
			enc,
		); err != nil {
			return ctx, e(err, "")
		}

		fsnodeinfo, err = launch.CreateLocalFS(
			launch.CreateDefaultNodeInfo(params.NetworkID(), version), design.Storage.Base, enc)
		if err != nil {
			return ctx, e(err, "")
		}
	default:
		return ctx, e(err, "")
	}

	ctx = context.WithValue(ctx, launch.FSNodeInfoContextKey, fsnodeinfo) //revive:disable-line:modifies-parameter

	return ctx, nil
}

func PLoadDatabase(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to load database")

	var log *logging.Logging
	var design NodeDesign
	var encs *encoder.Encoders
	var enc encoder.Encoder
	var fsnodeinfo launch.NodeInfo

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.DesignContextKey, &design,
		launch.EncodersContextKey, &encs,
		launch.EncoderContextKey, &enc,
		launch.FSNodeInfoContextKey, &fsnodeinfo,
	); err != nil {
		return ctx, e(err, "")
	}

	st, db, perm, pool, err := launch.LoadDatabase(
		fsnodeinfo, design.Storage.Database.String(), design.Storage.Base, encs, enc)
	if err != nil {
		return ctx, e(err, "")
	}

	_ = db.SetLogging(log)
	//revive:disable:modifies-parameter
	ctx = context.WithValue(ctx, launch.LeveldbStorageContextKey, st)
	ctx = context.WithValue(ctx, launch.CenterDatabaseContextKey, db)
	ctx = context.WithValue(ctx, launch.PermanentDatabaseContextKey, perm)
	ctx = context.WithValue(ctx, launch.PoolDatabaseContextKey, pool)
	//revive:enable:modifies-parameter

	return ctx, nil
}

func PLoadFromDatabase(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to load some stuffs from database")

	var design NodeDesign
	var encs *encoder.Encoders
	var center isaac.Database

	if err := util.LoadFromContextOK(ctx,
		launch.DesignContextKey, &design,
		launch.EncodersContextKey, &encs,
		launch.CenterDatabaseContextKey, &center,
	); err != nil {
		return ctx, e(err, "")
	}

	// NOTE load from last voteproofs
	lvps := isaacstates.NewLastVoteproofsHandler()
	ctx = context.WithValue(ctx, launch.LastVoteproofsHandlerContextKey, lvps) //revive:disable-line:modifies-parameter

	var manifest base.Manifest
	var enc encoder.Encoder

	switch m, found, err := center.LastBlockMap(); {
	case err != nil:
		return ctx, e(err, "")
	case !found:
		return ctx, nil
	default:
		enc = encs.Find(m.Encoder())
		if enc == nil {
			return ctx, e(nil, "encoder of last blockmap not found")
		}

		manifest = m.Manifest()
	}

	reader, err := isaacblock.NewLocalFSReaderFromHeight(
		launch.LocalFSDataDirectory(design.Storage.Base), manifest.Height(), enc,
	)
	if err != nil {
		return ctx, e(err, "")
	}

	defer func() {
		_ = reader.Close()
	}()

	switch v, found, err := reader.Item(base.BlockMapItemTypeVoteproofs); {
	case err != nil:
		return ctx, e(err, "")
	case !found:
		return ctx, e(nil, "last voteproofs not found in localfs")
	default:
		vps := v.([]base.Voteproof) //nolint:forcetypeassert //...

		lvps.Set(vps[0].(base.INITVoteproof))   //nolint:forcetypeassert //...
		lvps.Set(vps[1].(base.ACCEPTVoteproof)) //nolint:forcetypeassert //...
	}

	return ctx, nil
}

func PCleanStorage(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to clean storage")

	var design NodeDesign
	var encs *encoder.Encoders
	var enc encoder.Encoder

	if err := util.LoadFromContextOK(ctx,
		launch.DesignContextKey, &design,
		launch.EncodersContextKey, &encs,
		launch.EncoderContextKey, &enc,
	); err != nil {
		return ctx, e(err, "")
	}

	if err := launch.CleanStorage(design.Storage.Database.String(), design.Storage.Base, encs, enc); err != nil {
		return ctx, e(err, "")
	}

	return ctx, nil
}

func PCreateLocalFS(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to create localfs")

	var design NodeDesign
	var enc encoder.Encoder
	var params *isaac.LocalParams
	var version util.Version

	if err := util.LoadFromContextOK(ctx,
		launch.DesignContextKey, &design,
		launch.EncoderContextKey, &enc,
		launch.LocalParamsContextKey, &params,
		launch.VersionContextKey, &version,
	); err != nil {
		return ctx, e(err, "")
	}

	fsnodeinfo, err := launch.CreateLocalFS(
		launch.CreateDefaultNodeInfo(params.NetworkID(), version), design.Storage.Base, enc)
	if err != nil {
		return ctx, e(err, "")
	}

	ctx = context.WithValue(ctx, launch.FSNodeInfoContextKey, fsnodeinfo) //revive:disable-line:modifies-parameter

	return ctx, nil
}

func PNodeInfo(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to prepare nodeinfo")

	var log *logging.Logging
	var version util.Version
	var local base.LocalNode
	var params *isaac.LocalParams
	var design NodeDesign
	var db isaac.Database

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.VersionContextKey, &version,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return ctx, e(err, "")
	}

	nodeinfo := isaacnetwork.NewNodeInfoUpdater(design.NetworkID, local, version)
	_ = nodeinfo.SetConsensusState(isaacstates.StateBooting)
	_ = nodeinfo.SetConnInfo(network.ConnInfoToString(
		design.Network.PublishString,
		design.Network.TLSInsecure,
	))
	_ = nodeinfo.SetLocalParams(params)

	ctx = context.WithValue(ctx, launch.NodeInfoContextKey, nodeinfo) //revive:disable-line:modifies-parameter

	if err := launch.UpdateNodeInfoWithNewBlock(db, nodeinfo); err != nil {
		log.Log().Error().Err(err).Msg("failed to update nodeinfo")
	}

	return ctx, nil
}

func PNetwork(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to prepare network")

	var log *logging.Logging
	var encs *encoder.Encoders
	var enc encoder.Encoder
	var design NodeDesign
	var params base.LocalParams

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.EncodersContextKey, &encs,
		launch.EncoderContextKey, &enc,
		launch.DesignContextKey, &design,
		launch.LocalParamsContextKey, &params,
	); err != nil {
		return ctx, e(err, "")
	}

	handlers := quicstream.NewPrefixHandler(isaacnetwork.QuicstreamErrorHandler(enc))

	quicconfig := launch.DefaultQuicConfig()
	quicconfig.RequireAddressValidation = func(net.Addr) bool {
		return true // TODO NOTE manage blacklist
	}

	server := quicstream.NewServer(
		design.Network.Bind,
		launch.GenerateNewTLSConfig(params.NetworkID()),
		quicconfig,
		handlers.Handler,
	)
	_ = server.SetLogging(log)

	ctx = context.WithValue(ctx, launch.QuicstreamServerContextKey, server)     //revive:disable-line:modifies-parameter
	ctx = context.WithValue(ctx, launch.QuicstreamHandlersContextKey, handlers) //revive:disable-line:modifies-parameter

	return ctx, nil
}

func PSyncSourceChecker(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to prepare SyncSourceChecker")

	var log *logging.Logging
	var enc encoder.Encoder
	var design NodeDesign
	var local base.LocalNode
	var params *isaac.LocalParams
	var client *isaacnetwork.QuicstreamClient

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.EncoderContextKey, &enc,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		launch.QuicstreamClientContextKey, &client,
	); err != nil {
		return ctx, e(err, "")
	}

	sources := make([]isaacnetwork.SyncSource, len(design.SyncSources))
	copy(sources, design.SyncSources)

	switch {
	case len(sources) < 1:
		log.Log().Warn().Msg("empty initial sync sources; connected memberlist members will be used")
	default:
		log.Log().Debug().Interface("sync_sources", sources).Msg("initial sync sources found")
	}

	syncSourcePool := isaac.NewSyncSourcePool(nil)

	syncSourceChecker := isaacnetwork.NewSyncSourceChecker(
		local,
		params.NetworkID(),
		client,
		params.SyncSourceCheckerInterval(),
		enc,
		sources,
		func(called int64, ncis []isaac.NodeConnInfo, err error) {
			syncSourcePool.UpdateFixed(ncis)

			if err != nil {
				log.Log().Error().Err(err).
					Interface("node_conninfo", ncis).
					Msg("failed to check sync sources")

				return
			}

			log.Log().Debug().
				Int64("called", called).
				Interface("node_conninfo", ncis).
				Msg("sync sources updated")
		},
	)
	_ = syncSourceChecker.SetLogging(log)

	ctx = context.WithValue(ctx, //revive:disable-line:modifies-parameter
		launch.SyncSourceCheckerContextKey, syncSourceChecker)
	ctx = context.WithValue(ctx, //revive:disable-line:modifies-parameter
		launch.SyncSourcePoolContextKey, syncSourcePool)

	return ctx, nil
}

func PMemberlist(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to prepare memberlist")

	var log *logging.Logging
	var enc *jsonenc.Encoder
	var params *isaac.LocalParams

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.EncoderContextKey, &enc,
		launch.LocalParamsContextKey, &params,
	); err != nil {
		return ctx, e(err, "")
	}

	poolclient := quicstream.NewPoolClient()

	localnode, err := memberlistLocalNode(ctx)
	if err != nil {
		return ctx, e(err, "")
	}

	config, err := memberlistConfig(ctx, localnode, poolclient)
	if err != nil {
		return ctx, e(err, "")
	}

	m, err := quicmemberlist.NewMemberlist(
		localnode,
		enc,
		config,
		params.SameMemberLimit(),
	)
	if err != nil {
		return ctx, e(err, "")
	}

	_ = m.SetLogging(log)

	pps := ps.NewPS("event-on-empty-members")

	m.SetWhenLeftFunc(func(quicmemberlist.Node) {
		if m.IsJoined() {
			return
		}

		if _, err := pps.Run(context.Background()); err != nil {
			log.Log().Error().Err(err).Msg("failed to run onEmptyMembers")
		}
	})

	//revive:disable:modifies-parameter
	ctx = context.WithValue(ctx, launch.MemberlistContextKey, m)
	ctx = context.WithValue(ctx, launch.EventOnEmptyMembersContextKey, pps)
	//revive:enable:modifies-parameter

	return ctx, nil
}

func memberlistAlive(ctx context.Context) (*quicmemberlist.AliveDelegate, error) {
	var design NodeDesign
	var enc *jsonenc.Encoder

	if err := util.LoadFromContextOK(ctx,
		launch.DesignContextKey, &design,
		launch.EncoderContextKey, &enc,
	); err != nil {
		return nil, err
	}

	nc, err := launch.NodeChallengeFunc(ctx)
	if err != nil {
		return nil, err
	}

	al, err := launch.MemberlistAllowFunc(ctx)
	if err != nil {
		return nil, err
	}

	return quicmemberlist.NewAliveDelegate(
		enc,
		design.Network.Publish(),
		nc,
		al,
	), nil
}

func memberlistLocalNode(ctx context.Context) (quicmemberlist.Node, error) {
	var design NodeDesign
	var local base.LocalNode
	var fsnodeinfo launch.NodeInfo

	if err := util.LoadFromContextOK(ctx,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.FSNodeInfoContextKey, &fsnodeinfo,
	); err != nil {
		return nil, err
	}

	return quicmemberlist.NewNode(
		fsnodeinfo.ID(),
		design.Network.Publish(),
		local.Address(),
		local.Publickey(),
		design.Network.PublishString,
		design.Network.TLSInsecure,
	)
}

func memberlistConfig(
	ctx context.Context,
	localnode quicmemberlist.Node,
	poolclient *quicstream.PoolClient,
) (*memberlist.Config, error) {
	var log *logging.Logging
	var enc *jsonenc.Encoder
	var design NodeDesign
	var fsnodeinfo launch.NodeInfo
	var client *isaacnetwork.QuicstreamClient
	var local base.LocalNode
	var syncSourcePool *isaac.SyncSourcePool

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.EncoderContextKey, &enc,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.FSNodeInfoContextKey, &fsnodeinfo,
		launch.QuicstreamClientContextKey, &client,
		launch.EncoderContextKey, &enc,
		launch.LocalContextKey, &local,
		launch.SyncSourcePoolContextKey, &syncSourcePool,
	); err != nil {
		return nil, err
	}

	transport, err := memberlistTransport(ctx, poolclient)
	if err != nil {
		return nil, err
	}

	delegate := quicmemberlist.NewDelegate(localnode, nil, func(b []byte) {
		panic("set notifyMsgFunc")
	})

	alive, err := memberlistAlive(ctx)
	if err != nil {
		return nil, err
	}

	config := quicmemberlist.BasicMemberlistConfig(
		localnode.Name(),
		design.Network.Bind,
		design.Network.Publish(),
	)

	config.Transport = transport
	config.Delegate = delegate
	config.Alive = alive

	config.Events = quicmemberlist.NewEventsDelegate(
		enc,
		func(node quicmemberlist.Node) {
			l := log.Log().With().Interface("node", node).Logger()

			l.Debug().Msg("new node found")

			cctx, cancel := context.WithTimeout(
				context.Background(), time.Second*5) //nolint:gomnd //....
			defer cancel()

			c := client.NewQuicstreamClient(node.UDPConnInfo())(node.UDPAddr())
			if _, err := c.Dial(cctx); err != nil {
				l.Error().Err(err).Msg("new node joined, but failed to dial")

				return
			}

			poolclient.Add(node.UDPAddr(), c)

			if !node.Address().Equal(local.Address()) {
				nci := isaacnetwork.NewNodeConnInfoFromMemberlistNode(node)
				added := syncSourcePool.AddNonFixed(nci)

				l.Debug().
					Bool("added", added).
					Interface("node_conninfo", nci).
					Msg("new node added to SyncSourcePool")
			}
		},
		func(node quicmemberlist.Node) {
			l := log.Log().With().Interface("node", node).Logger()

			l.Debug().Msg("node left")

			if poolclient.Remove(node.UDPAddr()) {
				l.Debug().Msg("node removed from client pool")
			}

			nci := isaacnetwork.NewNodeConnInfoFromMemberlistNode(node)
			if syncSourcePool.RemoveNonFixed(nci) {
				l.Debug().Msg("node removed from sync source pool")
			}
		},
	)

	return config, nil
}

func memberlistTransport(
	ctx context.Context,
	poolclient *quicstream.PoolClient,
) (*quicmemberlist.Transport, error) {
	var log *logging.Logging
	var enc encoder.Encoder
	var design NodeDesign
	var fsnodeinfo launch.NodeInfo
	var client *isaacnetwork.QuicstreamClient
	var local base.LocalNode
	var syncSourcePool *isaac.SyncSourcePool
	var handlers *quicstream.PrefixHandler

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.EncoderContextKey, &enc,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.FSNodeInfoContextKey, &fsnodeinfo,
		launch.QuicstreamClientContextKey, &client,
		launch.EncoderContextKey, &enc,
		launch.LocalContextKey, &local,
		launch.SyncSourcePoolContextKey, &syncSourcePool,
		launch.QuicstreamHandlersContextKey, &handlers,
	); err != nil {
		return nil, err
	}

	transport := quicmemberlist.NewTransportWithQuicstream(
		design.Network.Publish(),
		isaacnetwork.HandlerPrefixMemberlist,
		poolclient,
		client.NewQuicstreamClient,
		nil,
	)

	_ = handlers.Add(isaacnetwork.HandlerPrefixMemberlist, func(addr net.Addr, r io.Reader, w io.Writer) error {
		b, err := io.ReadAll(r)
		if err != nil {
			log.Log().Error().Err(err).Stringer("remote_address", addr).Msg("failed to read")

			return errors.WithStack(err)
		}

		if err := transport.ReceiveRaw(b, addr); err != nil {
			log.Log().Error().Err(err).Stringer("remote_address", addr).Msg("invalid message received")

			return err
		}

		return nil
	})

	return transport, nil
}

func PCallbackBroadcaster(ctx context.Context) (context.Context, error) {
	var design NodeDesign
	var enc *jsonenc.Encoder
	var m *quicmemberlist.Memberlist

	if err := util.LoadFromContextOK(ctx,
		launch.DesignContextKey, &design,
		launch.EncoderContextKey, &enc,
		launch.MemberlistContextKey, &m,
	); err != nil {
		return nil, err
	}

	c := isaacnetwork.NewCallbackBroadcaster(
		quicstream.NewUDPConnInfo(design.Network.Publish(), design.Network.TLSInsecure),
		enc,
		m,
	)

	return context.WithValue(ctx, launch.CallbackBroadcasterContextKey, c), nil
}

func PNetworkHandlers(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to prepare network handlers")

	var encs *encoder.Encoders
	var enc encoder.Encoder
	var design NodeDesign
	var local base.LocalNode
	var params *isaac.LocalParams
	var db isaac.Database
	var pool *isaacdatabase.TempPool
	var proposalMaker *isaac.ProposalMaker
	var memberlist *quicmemberlist.Memberlist
	var syncSourcePool *isaac.SyncSourcePool
	var handlers *quicstream.PrefixHandler
	var nodeinfo *isaacnetwork.NodeInfoUpdater
	var svvotef isaac.SuffrageVoteFunc
	var cb *isaacnetwork.CallbackBroadcaster

	if err := util.LoadFromContextOK(ctx,
		launch.EncodersContextKey, &encs,
		launch.EncoderContextKey, &enc,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
		launch.PoolDatabaseContextKey, &pool,
		launch.ProposalMakerContextKey, &proposalMaker,
		launch.MemberlistContextKey, &memberlist,
		launch.SyncSourcePoolContextKey, &syncSourcePool,
		launch.QuicstreamHandlersContextKey, &handlers,
		launch.NodeInfoContextKey, &nodeinfo,
		launch.SuffrageVotingVoteFuncContextKey, &svvotef,
		launch.CallbackBroadcasterContextKey, &cb,
	); err != nil {
		return ctx, e(err, "")
	}

	sendOperationFilterf, err := SendOperationFilterFunc(ctx)
	if err != nil {
		return ctx, e(err, "")
	}

	idletimeout := time.Second * 2 //nolint:gomnd //...
	lastBlockMapf := launch.QuicstreamHandlerLastBlockMapFunc(db)
	suffrageNodeConnInfof := launch.QuicstreamHandlerSuffrageNodeConnInfoFunc(db, memberlist)

	handlers.
		Add(isaacnetwork.HandlerPrefixOperation, isaacnetwork.QuicstreamHandlerOperation(encs, idletimeout, pool)).
		Add(isaacnetwork.HandlerPrefixSendOperation,
			isaacnetwork.QuicstreamHandlerSendOperation(
				encs, idletimeout, params, pool,
				db.ExistsInStateOperation,
				sendOperationFilterf,
				svvotef,
				func(id string, b []byte) error {
					return cb.Broadcast(id, b, nil)
				},
			),
		).
		Add(isaacnetwork.HandlerPrefixRequestProposal,
			isaacnetwork.QuicstreamHandlerRequestProposal(encs, idletimeout,
				local, pool, proposalMaker, db.LastBlockMap,
			),
		).
		Add(isaacnetwork.HandlerPrefixProposal,
			isaacnetwork.QuicstreamHandlerProposal(encs, idletimeout, pool),
		).
		Add(isaacnetwork.HandlerPrefixLastSuffrageProof,
			isaacnetwork.QuicstreamHandlerLastSuffrageProof(encs, idletimeout,
				func(last util.Hash) (hint.Hint, []byte, []byte, bool, error) {
					enchint, metabytes, body, found, err := db.LastSuffrageProofBytes()

					switch {
					case err != nil:
						return enchint, nil, nil, false, err
					case !found:
						return enchint, nil, nil, false, storage.ErrNotFound.Errorf("last SuffrageProof not found")
					}

					switch h, err := isaacdatabase.ReadHashRecordMeta(metabytes); {
					case err != nil:
						return enchint, nil, nil, true, err
					case last != nil && last.Equal(h):
						return enchint, nil, nil, false, nil
					default:
						return enchint, metabytes, body, true, nil
					}
				},
			),
		).
		Add(isaacnetwork.HandlerPrefixSuffrageProof,
			isaacnetwork.QuicstreamHandlerSuffrageProof(encs, idletimeout, db.SuffrageProofBytes),
		).
		Add(isaacnetwork.HandlerPrefixLastBlockMap,
			isaacnetwork.QuicstreamHandlerLastBlockMap(encs, idletimeout, lastBlockMapf),
		).
		Add(isaacnetwork.HandlerPrefixBlockMap,
			isaacnetwork.QuicstreamHandlerBlockMap(encs, idletimeout, db.BlockMapBytes),
		).
		Add(isaacnetwork.HandlerPrefixBlockMapItem,
			isaacnetwork.QuicstreamHandlerBlockMapItem(encs, idletimeout, idletimeout*2, //nolint:gomnd //...
				func(height base.Height, item base.BlockMapItemType) (io.ReadCloser, bool, error) {
					e := util.StringErrorFunc("failed to get BlockMapItem")

					var menc encoder.Encoder

					switch m, found, err := db.BlockMap(height); {
					case err != nil:
						return nil, false, e(err, "")
					case !found:
						return nil, false, e(storage.ErrNotFound.Errorf("BlockMap not found"), "")
					default:
						menc = encs.Find(m.Encoder())
						if menc == nil {
							return nil, false, e(storage.ErrNotFound.Errorf("encoder of BlockMap not found"), "")
						}
					}

					reader, err := isaacblock.NewLocalFSReaderFromHeight(
						launch.LocalFSDataDirectory(design.Storage.Base), height, menc,
					)
					if err != nil {
						return nil, false, e(err, "")
					}
					defer func() {
						_ = reader.Close()
					}()

					return reader.Reader(item)
				},
			),
		).
		Add(isaacnetwork.HandlerPrefixNodeChallenge,
			isaacnetwork.QuicstreamHandlerNodeChallenge(encs, idletimeout, local, params),
		).
		Add(isaacnetwork.HandlerPrefixSuffrageNodeConnInfo,
			isaacnetwork.QuicstreamHandlerSuffrageNodeConnInfo(encs, idletimeout, suffrageNodeConnInfof),
		).
		Add(isaacnetwork.HandlerPrefixSyncSourceConnInfo,
			isaacnetwork.QuicstreamHandlerSyncSourceConnInfo(encs, idletimeout,
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
			),
		).
		Add(isaacnetwork.HandlerPrefixState,
			isaacnetwork.QuicstreamHandlerState(encs, idletimeout, db.StateBytes),
		).
		Add(isaacnetwork.HandlerPrefixExistsInStateOperation,
			isaacnetwork.QuicstreamHandlerExistsInStateOperation(encs, idletimeout, db.ExistsInStateOperation),
		).
		Add(isaacnetwork.HandlerPrefixNodeInfo,
			isaacnetwork.QuicstreamHandlerNodeInfo(encs, idletimeout, launch.QuicstreamHandlerGetNodeInfoFunc(enc, nodeinfo)),
		).
		Add(isaacnetwork.HandlerPrefixCallbackMessage,
			isaacnetwork.QuicstreamHandlerCallbackMessage(encs, idletimeout, cb),
		).
		Add(launch.HandlerPrefixPprof, launch.NetworkHandlerPprofFunc(encs))

	return ctx, nil
}

func SendOperationFilterFunc(ctx context.Context) (
	func(base.Operation) (bool, error),
	error,
) {
	var db isaac.Database
	var oprs *hint.CompatibleSet

	if err := util.LoadFromContextOK(ctx,
		launch.CenterDatabaseContextKey, &db,
		launch.OperationProcessorsMapContextKey, &oprs,
	); err != nil {
		return nil, err
	}

	operationfilterf := IsSupportedProposalOperationFactHintFunc()

	return func(op base.Operation) (bool, error) {
		switch hinter, ok := op.Fact().(hint.Hinter); {
		case !ok:
			return false, nil
		case !operationfilterf(hinter.Hint()):
			return false, nil
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

		f, closef, err := launch.OperationPreProcess(oprs, op, height)
		if err != nil {
			return false, err
		}

		defer func() {
			_ = closef()
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

func PProposalProcessors(ctx context.Context) (context.Context, error) {
	var log *logging.Logging

	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, err
	}

	newProposalProcessorf, err := newProposalProcessorFunc(ctx)
	if err != nil {
		return ctx, err
	}

	getProposalf, err := launch.GetProposalFunc(ctx)
	if err != nil {
		return ctx, err
	}

	pps := isaac.NewProposalProcessors(newProposalProcessorf, getProposalf)
	_ = pps.SetLogging(log)

	ctx = context.WithValue(ctx, launch.ProposalProcessorsContextKey, pps) //revive:disable-line:modifies-parameter

	return ctx, nil
}

func newProposalProcessorFunc(pctx context.Context) (
	func(base.ProposalSignFact, base.Manifest) (isaac.ProposalProcessor, error),
	error,
) {
	var enc encoder.Encoder
	var design NodeDesign
	var local base.LocalNode
	var params base.LocalParams
	var db isaac.Database
	var oprs *hint.CompatibleSet

	if err := util.LoadFromContextOK(pctx,
		launch.EncoderContextKey, &enc,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
		launch.OperationProcessorsMapContextKey, &oprs,
	); err != nil {
		return nil, err
	}

	getProposalOperationFuncf, err := launch.GetProposalOperationFunc(pctx)
	if err != nil {
		return nil, err
	}

	return func(proposal base.ProposalSignFact, previous base.Manifest) (
		isaac.ProposalProcessor, error,
	) {
		return isaac.NewDefaultProposalProcessor(
			proposal,
			previous,
			launch.NewBlockWriterFunc(
				local,
				params.NetworkID(),
				launch.LocalFSDataDirectory(design.Storage.Base),
				enc,
				db,
			),
			db.State,
			getProposalOperationFuncf(proposal),
			func(height base.Height, ht hint.Hint) (base.OperationProcessor, error) {
				v := oprs.Find(ht)
				if v == nil {
					return nil, nil
				}

				f := v.(func(height base.Height) (base.OperationProcessor, error)) //nolint:forcetypeassert //...

				return f(height)
			},
		)
	}, nil
}

func PStatesSetHandlers(ctx context.Context) (context.Context, error) { //revive:disable-line:function-length
	e := util.StringErrorFunc("failed to set states handler")

	var log *logging.Logging
	var local base.LocalNode
	var params *isaac.LocalParams
	var db isaac.Database
	var states *isaacstates.States
	var nodeinfo *isaacnetwork.NodeInfoUpdater
	var proposalSelector *isaac.BaseProposalSelector
	var nodeInConsensusNodesf func(base.Node, base.Height) (base.Suffrage, bool, error)
	var ballotbox *isaacstates.Ballotbox
	var pool *isaacdatabase.TempPool
	var lvps *isaacstates.LastVoteproofsHandler
	var pps *isaac.ProposalProcessors
	var sv *isaac.SuffrageVoting

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		launch.CenterDatabaseContextKey, &db,
		launch.StatesContextKey, &states,
		launch.NodeInfoContextKey, &nodeinfo,
		launch.ProposalSelectorContextKey, &proposalSelector,
		launch.NodeInConsensusNodesFuncContextKey, &nodeInConsensusNodesf,
		launch.BallotboxContextKey, &ballotbox,
		launch.PoolDatabaseContextKey, &pool,
		launch.LastVoteproofsHandlerContextKey, &lvps,
		launch.ProposalProcessorsContextKey, &pps,
		launch.SuffrageVotingContextKey, &sv,
	); err != nil {
		return ctx, e(err, "")
	}

	voteFunc := func(bl base.Ballot) (bool, error) {
		voted, err := ballotbox.Vote(bl, params.Threshold())
		if err != nil {
			return false, err
		}

		return voted, nil
	}

	joinMemberlistForStateHandlerf, err := launch.JoinMemberlistForStateHandlerFunc(ctx)
	if err != nil {
		return ctx, e(err, "")
	}

	joinMemberlistForJoiningeHandlerf, err := launch.JoinMemberlistForJoiningeHandlerFunc(ctx)
	if err != nil {
		return ctx, e(err, "")
	}

	leaveMemberlistForStateHandlerf, err := launch.LeaveMemberlistForStateHandlerFunc(ctx)
	if err != nil {
		return ctx, e(err, "")
	}

	leaveMemberlistForSyncingHandlerf, err := launch.LeaveMemberlistForSyncingHandlerFunc(ctx)
	if err != nil {
		return ctx, e(err, "")
	}

	var whenNewBlockSavedInSyncingStatef func(base.Height)

	switch err = util.LoadFromContext(
		ctx, launch.WhenNewBlockSavedInSyncingStateFuncContextKey, &whenNewBlockSavedInSyncingStatef); {
	case err != nil:
		return ctx, e(err, "")
	case whenNewBlockSavedInSyncingStatef == nil:
		whenNewBlockSavedInSyncingStatef = launch.WhenNewBlockSavedInSyncingStateFunc(db, nodeinfo)
	}

	var whenNewBlockSavedInConsensusStatef func(base.Height)

	switch err = util.LoadFromContext(
		ctx, launch.WhenNewBlockSavedInConsensusStateFuncContextKey, &whenNewBlockSavedInConsensusStatef); {
	case err != nil:
		return ctx, e(err, "")
	case whenNewBlockSavedInConsensusStatef == nil:
		whenNewBlockSavedInConsensusStatef = launch.WhenNewBlockSavedInConsensusStateFunc(params, ballotbox, db, nodeinfo)
	}

	suffrageVotingFindf := func(
		ctx context.Context,
		height base.Height,
		suf base.Suffrage,
	) ([]base.SuffrageWithdrawOperation, error) {
		return sv.Find(ctx, height, suf)
	}

	onEmptyMembersf, err := launch.OnEmptyMembersStateHandlerFunc(ctx, states)
	if err != nil {
		return ctx, e(err, "")
	}

	newsyncerf, err := newSyncerFunc(ctx, params, db, lvps, whenNewBlockSavedInSyncingStatef)
	if err != nil {
		return ctx, e(err, "")
	}

	getLastManifestf := launch.GetLastManifestFunc(db)
	getManifestf := launch.GetManifestFunc(db)

	whenSyncingFinished := func(base.Height) {
		ballotbox.Count(params.Threshold())
	}

	states.SetWhenStateSwitched(func(next isaacstates.StateType) {
		_ = nodeinfo.SetConsensusState(next)
	})

	syncinghandler := isaacstates.NewNewSyncingHandlerType(
		local, params, newsyncerf, nodeInConsensusNodesf,
		joinMemberlistForStateHandlerf,
		leaveMemberlistForSyncingHandlerf,
		whenNewBlockSavedInSyncingStatef,
	)
	syncinghandler.SetWhenFinished(whenSyncingFinished)

	consensusHandler := isaacstates.NewNewConsensusHandlerType(
		local, params, proposalSelector, pps,
		getManifestf, nodeInConsensusNodesf, voteFunc, whenNewBlockSavedInConsensusStatef, suffrageVotingFindf,
	)

	consensusHandler.SetOnEmptyMembers(onEmptyMembersf)

	joiningHandler := isaacstates.NewNewJoiningHandlerType(
		local, params, proposalSelector,
		getLastManifestf, nodeInConsensusNodesf,
		voteFunc, joinMemberlistForJoiningeHandlerf, leaveMemberlistForStateHandlerf, suffrageVotingFindf,
	)
	joiningHandler.SetOnEmptyMembers(onEmptyMembersf)

	states.
		SetHandler(isaacstates.StateBroken, isaacstates.NewNewBrokenHandlerType(local, params)).
		SetHandler(isaacstates.StateStopped, isaacstates.NewNewStoppedHandlerType(local, params)).
		SetHandler(
			isaacstates.StateBooting,
			isaacstates.NewNewBootingHandlerType(local, params,
				getLastManifestf, nodeInConsensusNodesf),
		).
		SetHandler(isaacstates.StateJoining, joiningHandler).
		SetHandler(isaacstates.StateConsensus, consensusHandler).
		SetHandler(isaacstates.StateSyncing, syncinghandler)

	_ = states.SetLogging(log)

	return ctx, nil
}

func newSyncerFunc(
	pctx context.Context,
	params *isaac.LocalParams,
	db isaac.Database,
	lvps *isaacstates.LastVoteproofsHandler,
	whenNewBlockSavedInSyncingStatef func(base.Height),
) (
	func(height base.Height) (isaac.Syncer, error),
	error,
) {
	var encs *encoder.Encoders
	var enc encoder.Encoder
	var devflags launch.DevFlags
	var design NodeDesign
	var client *isaacnetwork.QuicstreamClient
	var st *leveldbstorage.Storage
	var perm isaac.PermanentDatabase
	var syncSourcePool *isaac.SyncSourcePool

	if err := util.LoadFromContextOK(pctx,
		launch.EncodersContextKey, &encs,
		launch.EncoderContextKey, &enc,
		launch.DevFlagsContextKey, &devflags,
		launch.DesignContextKey, &design,
		launch.QuicstreamClientContextKey, &client,
		launch.LeveldbStorageContextKey, &st,
		launch.PermanentDatabaseContextKey, &perm,
		launch.SyncSourcePoolContextKey, &syncSourcePool,
	); err != nil {
		return nil, err
	}

	setLastVoteproofsfFromBlockReaderf, err := launch.SetLastVoteproofsfFromBlockReaderFunc(lvps)
	if err != nil {
		return nil, err
	}

	newSyncerDeferredf, err := launch.NewSyncerDeferredFunc(pctx, db)
	if err != nil {
		return nil, err
	}

	return func(height base.Height) (isaac.Syncer, error) {
		e := util.StringErrorFunc("failed newSyncer")

		var prev base.BlockMap

		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return nil, e(isaacstates.ErrUnpromising.Wrap(err), "")
		case found:
			prev = m
		}

		var tempsyncpool isaac.TempSyncPool

		switch i, err := isaacdatabase.NewLeveldbTempSyncPool(height, st, encs, enc); {
		case err != nil:
			return nil, e(isaacstates.ErrUnpromising.Wrap(err), "")
		default:
			tempsyncpool = i
		}

		newclient := client.Clone()

		var cachesize int64 = 333

		if prev != nil {
			if cachesize = (height - prev.Manifest().Height()).Int64(); cachesize > 333 { //nolint:gomnd //...
				cachesize = 333
			}
		}

		conninfocache, _ := util.NewShardedMap(base.NilHeight, quicstream.UDPConnInfo{}, 1<<9) //nolint:gomnd //...

		syncer, err := isaacstates.NewSyncer(
			design.Storage.Base,
			func(height base.Height) (isaac.BlockWriteDatabase, func(context.Context) error, error) {
				bwdb, err := db.NewBlockWriteDatabase(height)
				if err != nil {
					return nil, nil, err
				}

				return bwdb,
					func(ctx context.Context) error {
						if err := launch.MergeBlockWriteToPermanentDatabase(ctx, bwdb, perm); err != nil {
							return err
						}

						whenNewBlockSavedInSyncingStatef(height)

						return nil
					},
					nil
			},
			func(root string, blockmap base.BlockMap, bwdb isaac.BlockWriteDatabase) (isaac.BlockImporter, error) {
				return isaacblock.NewBlockImporter(
					launch.LocalFSDataDirectory(root),
					encs,
					blockmap,
					bwdb,
					params.NetworkID(),
				)
			},
			prev,
			launch.SyncerLastBlockMapFunc(newclient, params, syncSourcePool),
			launch.SyncerBlockMapFunc(newclient, params, syncSourcePool, conninfocache, devflags.DelaySyncer),
			launch.SyncerBlockMapItemFunc(newclient, conninfocache),
			tempsyncpool,
			setLastVoteproofsfFromBlockReaderf,
			func() error {
				conninfocache.Close()
				conninfocache = nil

				return newclient.Close()
			},
		)
		if err != nil {
			return nil, e(err, "")
		}

		go newSyncerDeferredf(height, syncer)

		return syncer, nil
	}, nil
}

func PWatchDesign(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to watch design")

	var log *logging.Logging
	var flag launch.DesignFlag

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.DesignFlagContextKey, &flag,
	); err != nil {
		return ctx, e(err, "")
	}

	watchUpdatefs, err := WatchUpdateFuncs(ctx)
	if err != nil {
		return ctx, e(err, "")
	}

	switch flag.Scheme() {
	case "consul":
		runf, err := launch.ConsulWatch(ctx, watchUpdatefs)
		if err != nil {
			return ctx, e(err, "failed to watch thru consul")
		}

		go func() {
			if err := runf(); err != nil {
				log.Log().Error().Err(err).Msg("watch stopped")
			}
		}()
	default:
		log.Log().Debug().Msg("design uri does not support watch")

		return ctx, nil
	}

	return ctx, nil
}

func WatchUpdateFuncs(ctx context.Context) (map[string]func(string) error, error) {
	var log *logging.Logging
	var enc *jsonenc.Encoder
	var design NodeDesign
	var params *isaac.LocalParams
	var discoveries *util.Locked[[]quicstream.UDPConnInfo]
	var syncSourceChecker *isaacnetwork.SyncSourceChecker

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.EncoderContextKey, &enc,
		launch.DesignContextKey, &design,
		launch.LocalParamsContextKey, &params,
		launch.DiscoveryContextKey, &discoveries,
		launch.SyncSourceCheckerContextKey, &syncSourceChecker,
	); err != nil {
		return nil, err
	}

	//revive:disable:line-length-limit
	updaters := map[string]func(string) error{
		"parameters/threshold":                                 launch.UpdateLocalParamThreshold(params, log),
		"parameters/interval_broadcast_ballot":                 launch.UpdateLocalParamIntervalBroadcastBallot(params, log),
		"parameters/wait_preparing_init_ballot":                launch.UpdateLocalParamWaitPreparingINITBallot(params, log),
		"parameters/timeout_request_proposal":                  launch.UpdateLocalParamTimeoutRequestProposal(params, log),
		"parameters/sync_source_checker_interval":              launch.UpdateLocalParamSyncSourceCheckerInterval(params, log),
		"parameters/valid_proposal_operation_expire":           launch.UpdateLocalParamValidProposalOperationExpire(params, log),
		"parameters/valid_proposal_suffrage_operations_expire": launch.UpdateLocalParamValidProposalSuffrageOperationsExpire(params, log),
		"parameters/max_operation_size":                        launch.UpdateLocalParamMaxOperationSize(params, log),
		"parameters/same_member_limit":                         launch.UpdateLocalParamSameMemberLimit(params, log),
		"discoveries":                                          launch.UpdateDiscoveries(discoveries, log),
		"sync_sources":                                         UpdateSyncSources(enc, design, syncSourceChecker, log),
	}
	//revive:enable:line-length-limit

	return updaters, nil
}

func UpdateSyncSources(
	enc *jsonenc.Encoder,
	design NodeDesign,
	syncSourceChecker *isaacnetwork.SyncSourceChecker,
	log *logging.Logging,
) func(string) error {
	return func(s string) error {
		e := util.StringErrorFunc("failed to update sync source")

		var sources launch.SyncSourcesDesign
		if err := sources.DecodeYAML([]byte(s), enc); err != nil {
			return e(err, "")
		}

		if err := launch.IsValidSyncSourcesDesign(
			sources,
			design.Address,
			design.Network.PublishString,
			design.Network.Publish().String(),
		); err != nil {
			return e(err, "")
		}

		prev := syncSourceChecker.Sources()
		syncSourceChecker.UpdateSources(sources)

		log.Log().Debug().
			Str("key", "sync_sources").
			Interface("prev", prev).
			Interface("updated", sources).
			Msg("sync sources updated")

		return nil
	}
}

func PStartTimeSyncer(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to prepare time syncer")

	var log *logging.Logging
	var design NodeDesign

	if err := util.LoadFromContextOK(ctx,
		launch.LoggingContextKey, &log,
		launch.DesignContextKey, &design,
	); err != nil {
		return ctx, e(err, "")
	}

	if len(design.TimeServer) < 1 {
		log.Log().Debug().Msg("no time server given")

		return ctx, nil
	}

	ts, err := localtime.NewTimeSyncer(design.TimeServer, design.TimeServerPort, launch.DefaultTimeSyncerInterval)
	if err != nil {
		return ctx, e(err, "")
	}

	_ = ts.SetLogging(log)

	if err := ts.Start(); err != nil {
		return ctx, e(err, "")
	}

	return context.WithValue(ctx, launch.TimeSyncerContextKey, ts), nil
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
