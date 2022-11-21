package cmds

type KeyCommand struct {
	New     KeyNewCommand     `cmd:"" help:"generate new key"`
	Address KeyAddressCommand `cmd:"" help:"generate address from key"`
	Load    KeyLoadCommand    `cmd:"" help:"load key"`
	Sign    KeySignCommand    `cmd:"" help:"sign"`
}

func NewKeyCommand() KeyCommand {
	return KeyCommand{
		New:     NewKeyNewCommand(),
		Address: NewKeyAddressCommand(),
		Load:    NewKeyLoadCommand(),
		Sign:    NewKeySignCommand(),
	}
}
