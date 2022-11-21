package cmds

type SealCommand struct {
	CreateAccount    CreateAccountCommand    `cmd:"" name:"create-account" help:"create new account"`
	Transfer         TransferCommand         `cmd:"" name:"transfer" help:"transfer"`
	CurrencyRegister CurrencyRegisterCommand `cmd:"" name:"currency-register" help:"register new currency"`
}

func NewSealCommand() SealCommand {
	return SealCommand{
		CreateAccount:    NewCreateAccountCommand(),
		Transfer:         NewTransferCommand(),
		CurrencyRegister: NewCurrencyRegisterCommand(),
	}
}
