package currency

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/logging"
)

var operationProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(OperationProcessor)
	},
}

type GetNewProcessor func(
	height base.Height,
	getStateFunc base.GetStateFunc,
	newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	newProcessConstraintFunc base.NewOperationProcessorProcessFunc) (base.OperationProcessor, error)

type DuplicationType string

type AddFee map[CurrencyID][2]Big

func (af AddFee) Fee(key CurrencyID, fee Big) AddFee {
	switch v, found := af[key]; {
	case !found:
		af[key] = [2]Big{ZeroBig, fee}
	default:
		af[key] = [2]Big{v[0], v[1].Add(fee)}
	}

	return af
}

func (af AddFee) Add(key CurrencyID, add Big) AddFee {
	switch v, found := af[key]; {
	case !found:
		af[key] = [2]Big{add, ZeroBig}
	default:
		af[key] = [2]Big{v[0].Add(add), v[1]}
	}

	return af
}

const (
	DuplicationTypeSender   DuplicationType = "sender"
	DuplicationTypeCurrency DuplicationType = "currency"
)

type BaseOperationProcessor interface {
	PreProcess(base.Operation, base.GetStateFunc) (base.OperationProcessReasonError, error)
	Process(base.Operation, base.GetStateFunc) ([]base.StateMergeValue, base.OperationProcessReasonError, error)
	Close() error
}

type OperationProcessor struct {
	// id string
	sync.RWMutex
	*logging.Logging
	*base.BaseOperationProcessor
	processorHintSet     *hint.CompatibleSet
	fee                  map[CurrencyID]Big
	duplicated           map[string]DuplicationType
	duplicatedNewAddress map[string]struct{}
	processorClosers     *sync.Map
	GetStateFunc         base.GetStateFunc
	CollectFee           func(*OperationProcessor, AddFee) error
}

/*
func NewOperationProcessor(
	height base.Height,
	getStateFunc base.GetStateFunc,
	newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
) (*OperationProcessor, error) {
	e := util.StringErrorFunc("failed to create new OperationProcessor")

	b, err := base.NewBaseOperationProcessor(
		height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
	if err != nil {
		return nil, e(err, "")
	}

	return &OperationProcessor{
		id: util.UUID().String(),
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "mitum-currency-operations-processor")
		}),
		BaseOperationProcessor: b,
		processorHintSet:       hint.NewCompatibleSet(),
		fee:                    map[CurrencyID]Big{},
		duplicated:             map[string]DuplicationType{},
		duplicatedNewAddress:   map[string]struct{}{},
		processorClosers:       &sync.Map{},
		GetStateFunc:           getStateFunc,
	}, nil
}
*/

func NewOperationProcessor() *OperationProcessor {
	m := sync.Map{}
	return &OperationProcessor{
		// id: util.UUID().String(),
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "mitum-currency-operations-processor")
		}),
		processorHintSet:     hint.NewCompatibleSet(),
		fee:                  map[CurrencyID]Big{},
		duplicated:           map[string]DuplicationType{},
		duplicatedNewAddress: map[string]struct{}{},
		processorClosers:     &m,
		CollectFee:           CollectFee,
	}
}

func (opr *OperationProcessor) New(
	height base.Height,
	getStateFunc base.GetStateFunc,
	newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	newProcessConstraintFunc base.NewOperationProcessorProcessFunc) (*OperationProcessor, error) {
	e := util.StringErrorFunc("failed to create new OperationProcessor")

	nopr := operationProcessorPool.Get().(*OperationProcessor)
	if nopr.processorHintSet == nil {
		nopr.processorHintSet = opr.processorHintSet
	}

	if nopr.fee == nil {
		nopr.fee = opr.fee
	}

	if nopr.duplicated == nil {
		nopr.duplicated = make(map[string]DuplicationType)
	}

	// if nopr.duplicatedNewAddress == nil {
	// 	nopr.duplicatedNewAddress = opr.duplicatedNewAddress
	// }

	if nopr.duplicatedNewAddress == nil {
		nopr.duplicatedNewAddress = make(map[string]struct{})
	}

	if nopr.Logging == nil {
		nopr.Logging = opr.Logging
	}

	b, err := base.NewBaseOperationProcessor(
		height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
	if err != nil {
		return nil, e(err, "")
	}

	nopr.BaseOperationProcessor = b
	nopr.GetStateFunc = getStateFunc
	return nopr, nil
}

func (opr *OperationProcessor) SetProcessor(
	hint hint.Hint,
	newProcessor GetNewProcessor,
) (base.OperationProcessor, error) {
	if err := opr.processorHintSet.Add(hint, newProcessor); err != nil {
		if !errors.Is(err, util.ErrFound) {
			return nil, err
		}
	}

	return opr, nil
}

func (opr *OperationProcessor) PreProcess(ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to preprocess for OperationProcessor")

	if opr.processorClosers == nil {
		opr.processorClosers = &sync.Map{}
	}

	var sp base.OperationProcessor
	switch i, known, err := opr.getNewProcessor(op); {
	case err != nil:
		return ctx, base.NewBaseOperationProcessReasonError(err.Error()), nil
	case !known:
		return ctx, nil, e(nil, "failed to getNewProcessor, %T", op)
	default:
		sp = i
	}

	switch _, reasonerr, err := sp.PreProcess(ctx, op, getStateFunc); {
	case err != nil:
		return ctx, nil, e(err, "")
	case reasonerr != nil:
		return ctx, reasonerr, nil
	}

	return ctx, nil, nil
}

func (opr *OperationProcessor) Process(ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to process for OperationProcessor")

	if err := opr.checkDuplication(op); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("duplication found: %w", err), nil
	}

	var sp base.OperationProcessor
	switch i, known, err := opr.getNewProcessor(op); {
	case err != nil:
		return nil, nil, e(err, "")
	case !known:
		return nil, nil, e(nil, "failed to getNewProcessor")
	default:
		sp = i
	}

	stateMergeValues, reasonerr, err := sp.Process(ctx, op, getStateFunc)
	// opr.collecFee(stateMergeValues...)

	return stateMergeValues, reasonerr, err
}

/*
func (opr *OperationProcessor) process(op base.OperationProcessor) error {
	var sp base.OperationProcessor

	switch t := op.(type) {
	case *TransfersProcessor:
		sp = t
	case *CreateAccountsProcessor:
		sp = t
	case *KeyUpdaterProcessor:
		sp = t
	default:
		return op.Process(opr.pool.Get, opr.pool.Set)
	}

	return sp.Process(opr.pool.Get, opr.setState)
}
*/

func (opr *OperationProcessor) checkDuplication(op base.Operation) error {
	opr.Lock()
	defer opr.Unlock()

	// var did string
	// var didtype DuplicationType
	var newAddresses []base.Address

	switch t := op.(type) {
	case CreateAccounts:
		fact, ok := t.Fact().(CreateAccountsFact)
		if !ok {
			return errors.Errorf("expected GenesisCurrenciesFact, not %T", t.Fact())
		}
		as, err := fact.Targets()
		if err != nil {
			return errors.Errorf("failed to get Addresses")
		}
		newAddresses = as
		// did = fact.Sender().String()
		// didtype = DuplicationTypeSender
	// case Transfers:
	// 	fact, ok := t.Fact().(TransfersFact)
	// 	if !ok {
	// 		return errors.Errorf("expected TransfersFact, not %T", t.Fact())
	// 	}
	// 	did = fact.Sender().String()
	// 	didtype = DuplicationTypeSender
	// case CurrencyRegister:
	// 	fact, ok := t.Fact().(CurrencyRegisterFact)
	// 	if !ok {
	// 		return errors.Errorf("expected CurrencyRegisterFact, not %T", t.Fact())
	// 	}
	// 	did = fact.Currency().amount.Currency().String()
	// 	didtype = DuplicationTypeCurrency
	default:
		return nil
	}

	// if len(did) > 0 {
	// 	if _, found := opr.duplicated[did]; found {
	// 		switch didtype {
	// 		case DuplicationTypeSender:
	// 			return errors.Errorf("violates only one sender in proposal")
	// 		case DuplicationTypeCurrency:
	// 			return errors.Errorf("duplicated currency id, %q found in proposal", did)
	// 		default:
	// 			return errors.Errorf("violates duplication in proposal")
	// 		}
	// 	}

	// 	opr.duplicated[did] = didtype
	// }

	if len(newAddresses) > 0 {
		if err := opr.checkNewAddressDuplication(newAddresses); err != nil {
			return err
		}
	}

	return nil
}

func (opr *OperationProcessor) checkNewAddressDuplication(as []base.Address) error {
	for i := range as {
		if _, found := opr.duplicatedNewAddress[as[i].String()]; found {
			return errors.Errorf("new address already processed")
		}
	}

	for i := range as {
		opr.duplicatedNewAddress[as[i].String()] = struct{}{}
	}

	return nil
}

func (opr *OperationProcessor) Close() error {
	opr.Lock()
	defer opr.Unlock()

	defer opr.close()
	/*
		if len(opr.fee) > 0 {
			op, err := NewFeeOperation(NewFeeOperationFact(opr.Height(), opr.fee))
			if err != nil {
				return err
			}

			pr, err := NewFeeOperationProcessor(opr.Height(), opr.GetStateFunc)
			if err != nil {
				return err
			}

				if err := pr.Process(nil, op, opr.GetStateFunc); err != nil {
					return err
				}
				opr.pool.AddOperations(op)

		}
	*/

	return nil
}

func (opr *OperationProcessor) Cancel() error {
	opr.Lock()
	defer opr.Unlock()

	defer opr.close()

	return nil
}

func (opr *OperationProcessor) getNewProcessor(op base.Operation) (base.OperationProcessor, bool, error) {
	switch i, err := opr.getNewProcessorFromHintset(op); {
	case err != nil:
		return nil, false, err
	case i != nil:
		return i, true, nil
	}

	switch t := op.(type) {
	case CreateAccounts,
		Transfers,
		CurrencyRegister:
		return nil, false, errors.Errorf("%T needs SetProcessor", t)
	default:
		return nil, false, nil
	}
}

func (opr *OperationProcessor) getNewProcessorFromHintset(op base.Operation) (base.OperationProcessor, error) {
	var f GetNewProcessor
	if hinter, ok := op.(hint.Hinter); !ok {
		return nil, nil
	} else if i := opr.processorHintSet.Find(hinter.Hint()); i == nil {
		return nil, nil
	} else if j, ok := i.(GetNewProcessor); !ok {
		return nil, errors.Errorf("invalid GetNewProcessor func, %T", i)
	} else {
		f = j
	}

	opp, err := f(opr.Height(), opr.GetStateFunc, nil, nil)
	if err != nil {
		return nil, err
	}

	h := op.(util.Hasher).Hash().String()
	_, iscloser := opp.(io.Closer)
	if iscloser {
		opr.processorClosers.Store(h, opp)
		iscloser = true
	}

	opr.Log().Debug().
		Str("operation", h).
		Str("processor", fmt.Sprintf("%T", opp)).
		Bool("is_closer", iscloser).
		Msg("operation processor created")

	return opp, nil
}

func (opr *OperationProcessor) close() {
	opr.processorClosers.Range(func(_, v interface{}) bool {
		err := v.(io.Closer).Close()
		if err != nil {
			opr.Log().Error().Err(err).Str("op", fmt.Sprintf("%T", v)).Msg("failed to close operation processor")
		} else {
			opr.Log().Debug().Str("processor", fmt.Sprintf("%T", v)).Msg("operation processor closed")
		}

		return true
	})

	// opr.pool = nil
	opr.fee = nil
	opr.duplicated = nil
	opr.duplicatedNewAddress = nil
	opr.processorClosers = &sync.Map{}

	operationProcessorPool.Put(opr)

	opr.Log().Debug().Msg("operation processors closed")
}

func CollectFee(opr *OperationProcessor, required AddFee) error {
	opr.Lock()
	defer opr.Unlock()

	for k, v := range required {
		if k == "" {
			return errors.Errorf("AddFee key is empty")
		}
		if v == [2]Big{} {
			return errors.Errorf("AddFee value is nil")
		}
		switch i, found := opr.fee[k]; {
		case !found:
			opr.fee[k] = v[1]
		default:
			opr.fee[k] = i.Add(v[1])
		}
	}
	return nil
}
