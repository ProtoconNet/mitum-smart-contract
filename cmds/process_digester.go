package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum2/base"
	isaacblock "github.com/ProtoconNet/mitum2/isaac/block"
	isaacdatabase "github.com/ProtoconNet/mitum2/isaac/database"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/fixedtree"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/ProtoconNet/mitum2/util/ps"
	"github.com/pkg/errors"
)

const (
	PNameDigester      = ps.Name("digester")
	PNameStartDigester = ps.Name("start_digester")
)

func ProcessDigester(ctx context.Context) (context.Context, error) {
	var log *logging.Logging
	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, err
	}

	var st *digest.Database
	if err := util.LoadFromContext(ctx, ContextValueDigestDatabase, &st); err != nil {
		return ctx, err
	}

	if st == nil {
		return ctx, nil
	}

	var design launch.NodeDesign
	if err := util.LoadFromContext(ctx, launch.DesignContextKey, &design); err != nil {
		return ctx, err
	}
	root := launch.LocalFSDataDirectory(design.Storage.Base)

	di := digest.NewDigester(st, root, nil)
	_ = di.SetLogging(log)

	return context.WithValue(ctx, ContextValueDigester, di), nil
}

func ProcessStartDigester(ctx context.Context) (context.Context, error) {
	var di *digest.Digester
	if err := util.LoadFromContext(ctx, ContextValueDigester, &di); err != nil {
		return ctx, err
	}
	if di == nil {
		return ctx, nil
	}

	return ctx, di.Start(ctx)
}

func PdigesterFollowUp(ctx context.Context) (context.Context, error) {
	var log *logging.Logging
	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, err
	}

	log.Log().Debug().Msg("digester trying to follow up")

	var mst *isaacdatabase.Center
	if err := util.LoadFromContextOK(ctx, launch.CenterDatabaseContextKey, &mst); err != nil {
		return ctx, err
	}

	var st *digest.Database
	if err := util.LoadFromContext(ctx, ContextValueDigestDatabase, &st); err != nil {
		return ctx, err
	}
	if st == nil {
		return ctx, nil
	}

	switch m, found, err := mst.LastBlockMap(); {
	case err != nil:
		return ctx, err
	case !found:
		log.Log().Debug().Msg("last BlockMap not found")
	case m.Manifest().Height() > st.LastBlock():
		log.Log().Info().
			Int64("last_manifest", m.Manifest().Height().Int64()).
			Int64("last_block", st.LastBlock().Int64()).
			Msg("new blocks found to digest")

		if err := digestFollowup(ctx, m.Manifest().Height()); err != nil {
			log.Log().Error().Err(err).Msg("follow up")

			return ctx, err
		}
		log.Log().Info().Msg("digested new blocks")
	default:
		log.Log().Info().Msg("digested blocks is up-to-dated")
	}

	return ctx, nil
}

func digestFollowup(ctx context.Context, height base.Height) error {
	var st *digest.Database
	if err := util.LoadFromContextOK(ctx, ContextValueDigestDatabase, &st); err != nil {
		return err
	}

	var design launch.NodeDesign
	if err := util.LoadFromContext(ctx, launch.DesignContextKey, &design); err != nil {
		return err
	}
	root := launch.LocalFSDataDirectory(design.Storage.Base)

	// var cp *currency.CurrencyPool
	// if err := LoadCurrencyPoolContextValue(ctx, &cp); err != nil {
	// 	return err
	// }

	if height <= st.LastBlock() {
		return nil
	}

	lastBlock := st.LastBlock()
	if lastBlock < base.GenesisHeight {
		lastBlock = base.GenesisHeight
	}

	for i := lastBlock; i <= height; i++ {
		reader, err := isaacblock.NewLocalFSReaderFromHeight(root, i, enc)
		if err != nil {
			return err
		}
		m, found, err := reader.BlockMap()
		if err != nil {
			return err
		} else if !found {
			return errors.Errorf("blockmap not found")
		}
		if err := m.IsValid(design.NetworkID); err != nil {
			return err
		}

		var ops []base.Operation
		switch v, found, err := reader.Item(base.BlockMapItemTypeOperations); {
		case err != nil:
			return err
		case found:
			ops = v.([]base.Operation) //nolint:forcetypeassert //...
		}

		var opstree fixedtree.Tree
		switch v, found, err := reader.Item(base.BlockMapItemTypeOperationsTree); {
		case err != nil:
			return err
		case found:
			opstree = v.(fixedtree.Tree) //nolint:forcetypeassert //...
		}

		var sts []base.State
		switch v, found, err := reader.Item(base.BlockMapItemTypeStates); {
		case err != nil:
			return err
		case found:
			sts = v.([]base.State) //nolint:forcetypeassert //...
		}

		if err := digest.DigestBlock(ctx, st, m, ops, opstree, sts); err != nil {
			return err
		}

		if err := st.SetLastBlock(m.Manifest().Height()); err != nil {
			return err
		}

	}
	return nil
}
