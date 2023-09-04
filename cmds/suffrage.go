package cmds

type SuffrageCommand struct {
	Mint              MintCommand              `cmd:"" name:"mint" help:"mint operation"`
	SuffrageCandidate SuffrageCandidateCommand `cmd:"" name:"suffrage-candidate" help:"suffrage candidate operation"`
	SuffrageJoin      SuffrageJoinCommand      `cmd:"" name:"suffrage-join" help:"suffrage join operation"`
	SuffrageDisjoin   SuffrageDisjoinCommand   `cmd:"" name:"suffrage-disjoin" help:"suffrage disjoin operation"` // revive:disable-line:line-length-limit
}
