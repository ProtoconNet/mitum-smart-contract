package cmds

import (
	"context"
	"github.com/ProtoconNet/mitum2/isaac"

	"github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum2/base"
	isaacblock "github.com/ProtoconNet/mitum2/isaac/block"
	isaacdatabase "github.com/ProtoconNet/mitum2/isaac/database"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/ProtoconNet/mitum2/util/ps"
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

	var newReaders func(context.Context, string, *isaac.BlockItemReadersArgs) (*isaac.BlockItemReaders, error)
	var fromRemotes isaac.RemotesBlockItemReadFunc

	if err := util.LoadFromContextOK(ctx,
		launch.NewBlockItemReadersFuncContextKey, &newReaders,
		launch.RemotesBlockItemReaderFuncContextKey, &fromRemotes,
	); err != nil {
		return ctx, err
	}

	var sourceReaders *isaac.BlockItemReaders

	switch i, err := newReaders(ctx, root, nil); {
	case err != nil:
		return ctx, err
	default:
		sourceReaders = i
	}

	di := digest.NewDigester(st, root, sourceReaders, fromRemotes, design.NetworkID, nil)
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

	var newReaders func(context.Context, string, *isaac.BlockItemReadersArgs) (*isaac.BlockItemReaders, error)
	var fromRemotes isaac.RemotesBlockItemReadFunc

	if err := util.LoadFromContextOK(ctx,
		launch.NewBlockItemReadersFuncContextKey, &newReaders,
		launch.RemotesBlockItemReaderFuncContextKey, &fromRemotes,
	); err != nil {
		return err
	}

	var sourceReaders *isaac.BlockItemReaders

	switch i, err := newReaders(ctx, root, nil); {
	case err != nil:
		return err
	default:
		sourceReaders = i
	}

	if height <= st.LastBlock() {
		return nil
	}

	lastBlock := st.LastBlock()
	if lastBlock < base.GenesisHeight {
		lastBlock = base.GenesisHeight
	}

	for h := lastBlock; h <= height; h++ {
		var bm base.BlockMap

		switch i, found, err := isaac.BlockItemReadersDecode[base.BlockMap](sourceReaders.Item, h, base.BlockItemMap, nil); {
		case err != nil:
			return err
		case !found:
			return util.ErrNotFound.Errorf("blockmap")
		default:
			if err := i.IsValid(design.NetworkID); err != nil {
				return err
			}

			bm = i
		}

		pr, ops, sts, opsTree, _, _, err := isaacblock.LoadBlockItemsFromReader(bm, sourceReaders.Item, h)
		if err != nil {
			return err
		}

		if err := digest.DigestBlock(ctx, st, bm, ops, opsTree, sts, pr); err != nil {
			return err
		}

		if err := st.SetLastBlock(h); err != nil {
			return err
		}

	}
	return nil
}
