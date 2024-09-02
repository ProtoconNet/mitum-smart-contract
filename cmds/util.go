package cmds

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
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
	"io"
	"os"
)

var (
	PNameDigestDesign                   = ps.Name("digest-design")
	PNameGenerateGenesis                = ps.Name("mitum-currency-generate-genesis")
	PNameDigestAPIHandlers              = ps.Name("mitum-currency-digest-api-handlers")
	PNameDigesterFollowUp               = ps.Name("mitum-currency-followup_digester")
	BEncoderContextKey                  = util.ContextKey("bson-encoder")
	ProposalOperationFactHintContextKey = util.ContextKey("proposal-operation-fact-hint")
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
		isaacnetwork.HandlerNameSendOperation,
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
	var oprsB *hint.CompatibleSet[NewOperationProcessorInternalWithProposalFunc]
	var f ProposalOperationFactHintFunc

	if err := util.LoadFromContextOK(ctx,
		launch.CenterDatabaseContextKey, &db,
		launch.OperationProcessorsMapContextKey, &oprs,
		OperationProcessorsMapBContextKey, &oprsB,
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

		f, closeF, err := OperationPreProcess(db, oprs, oprsB, op, height)
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

func OperationPreProcess(
	db isaac.Database,
	oprsA *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc],
	oprsB *hint.CompatibleSet[NewOperationProcessorInternalWithProposalFunc],
	op base.Operation,
	height base.Height,
) (
	preprocess func(context.Context, base.GetStateFunc) (context.Context, base.OperationProcessReasonError, error),
	cancel func() error,
	_ error,
) {
	fA, foundA := oprsA.Find(op.Hint())
	fB, foundB := oprsB.Find(op.Hint())
	if !foundA && !foundB {
		return op.PreProcess, util.EmptyCancelFunc, nil
	}

	if foundA {
		switch opp, err := fA(height, db.State); {
		case err != nil:
			return nil, nil, err
		default:
			return func(pctx context.Context, getStateFunc base.GetStateFunc) (
				context.Context, base.OperationProcessReasonError, error,
			) {
				return opp.PreProcess(pctx, op, getStateFunc)
			}, opp.Close, nil
		}
	}
	switch opp, err := fB(height, nil, db.State); {
	case err != nil:
		return nil, nil, err
	default:
		return func(pctx context.Context, getStateFunc base.GetStateFunc) (
			context.Context, base.OperationProcessReasonError, error,
		) {
			return opp.PreProcess(pctx, op, getStateFunc)
		}, opp.Close, nil
	}

}
