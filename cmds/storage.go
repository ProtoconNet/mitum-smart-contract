package cmds

import launchcmd "github.com/ProtoconNet/mitum2/launch/cmd"

type Storage struct { //nolint:govet //...
	Import         ImportCommand                   `cmd:"" help:"import block data files"`
	Clean          launchcmd.CleanCommand          `cmd:"" help:"clean storage"`
	ValidateBlocks launchcmd.ValidateBlocksCommand `cmd:"" help:"validate blocks in storage"`
	Status         launchcmd.StorageStatusCommand  `cmd:"" help:"storage status"`
	Database       launchcmd.DatabaseCommand       `cmd:"" help:""`
}
