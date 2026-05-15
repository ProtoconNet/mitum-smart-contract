package contract

import cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"

var contractEngine cruntime.ContractEngine = cruntime.NewHybridEngine(
	cruntime.NewYaegiEngine(),
	cruntime.NewGnoEngine(),
)
