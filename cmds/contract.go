package cmds

type SmartContractCommand struct {
	RegisterContractCommand RegisterContractCommand `cmd:"" name:"register-contract" help:"register contract"`
	CallContractCommand     CallContractCommand     `cmd:"" name:"call-contract" help:"call contract"`
}
