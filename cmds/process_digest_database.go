package cmds

import (
	"context"

	mongodbstorage "github.com/spikeekips/mitum-currency/digest/mongodb"
	isaacdatabase "github.com/spikeekips/mitum/isaac/database"
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"

	"github.com/spikeekips/mitum-currency/digest"
)

const ProcessNameDigestDatabase = "digest_database"

func ProcessDigestDatabase(ctx context.Context) (context.Context, error) {
	var design DigestDesign
	if err := util.LoadFromContext(ctx, ContextValueDigestDesign, &design); err != nil {
		return ctx, err
	}

	var mst *isaacdatabase.Center
	if err := util.LoadFromContextOK(ctx, launch.CenterDatabaseContextKey, &mst); err != nil {
		return ctx, err
	}

	dst, err := mongodbstorage.NewDatabaseFromURI(design.Database().URI().String(), encs)
	if err != nil {
		return ctx, err
	}

	st, err := loadDigestDatabase(mst, dst, false)
	if err != nil {
		return ctx, err
	}

	var log *logging.Logging
	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, err
	}

	_ = st.SetLogging(log)

	return context.WithValue(ctx, ContextValueDigestDatabase, st), nil
}

func loadDigestDatabase(mst *isaacdatabase.Center, st *mongodbstorage.Database, readonly bool) (*digest.Database, error) {
	var dst *digest.Database
	if readonly {
		s, err := digest.NewReadonlyDatabase(mst, st)
		if err != nil {
			return nil, err
		}
		dst = s
	} else {
		s, err := digest.NewDatabase(mst, st)
		if err != nil {
			return nil, err
		}
		dst = s
	}

	if err := dst.Initialize(); err != nil {
		return nil, err
	}

	return dst, nil
}
