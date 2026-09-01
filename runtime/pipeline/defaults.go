package pipeline

import (
	capi "github.com/imfact-labs/currency-model/api"
	ccmds "github.com/imfact-labs/currency-model/app/cmds"
	cpipeline "github.com/imfact-labs/currency-model/app/runtime/pipeline"
	csteps "github.com/imfact-labs/currency-model/app/runtime/steps"
	cdigest "github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/imfact-labs/smart-contract-model/runtime/steps"
)

type RunHooks struct {
	DigesterName             ps.Name
	Digester                 ps.Func
	StartDigesterName        ps.Name
	StartDigester            ps.Func
	CheckHold                ps.Func
	ProposalProcessors       ps.Func
	WhenNewBlockSaved        ps.Func
	WhenNewBlockConfirmed    ps.Func
	WhenNewBlockSavedSyncing ps.Func
	DigestAPIHandlers        ps.Func
	DigesterFollowUp         ps.Func
}

func DefaultRunPS(hooks RunHooks) *ps.PS {
	pps := cpipeline.DefaultRunPS()
	_ = pps.POK(launch.PNameStates).PreRemoveOK(launch.PNameProposalProcessors)
	_ = pps.AddOK(hooks.DigesterName, hooks.Digester, nil, cdigest.PNameDigesterDataBase).
		AddOK(hooks.StartDigesterName, hooks.StartDigester, nil, capi.PNameStartAPI)
	_ = pps.POK(launch.PNameStorage).PostAddOK(ps.Name("check-hold"), hooks.CheckHold)
	_ = pps.POK(launch.PNameStates).
		PreAddOK(steps.PNameOperationProcessorsMap, steps.POperationProcessorsMap).
		PreAddOK(launch.PNameProposalProcessors, hooks.ProposalProcessors).
		PreAddOK(ps.Name("when-new-block-saved-in-consensus-state-func"), hooks.WhenNewBlockSaved).
		PreAddOK(ps.Name("when-new-block-confirmed-func"), hooks.WhenNewBlockConfirmed).
		PreAddOK(ps.Name("when-new-block-saved-in-syncing-state-func"), hooks.WhenNewBlockSavedSyncing)
	_ = pps.POK(launch.PNameEncoder).PostAddOK(launch.PNameAddHinters, steps.PAddHinters)
	_ = pps.POK(capi.PNameAPI).PostAddOK(ccmds.PNameDigestAPIHandlers, hooks.DigestAPIHandlers)
	_ = pps.POK(hooks.DigesterName).PostAddOK(ccmds.PNameDigesterFollowUp, hooks.DigesterFollowUp)
	return pps
}

func DefaultImportPS() *ps.PS {
	pps := ps.NewPS("cmd-import")
	_ = pps.
		AddOK(launch.PNameEncoder, csteps.PEncoder, nil).
		AddOK(launch.PNameDesign, launch.PLoadDesign, nil, launch.PNameEncoder).
		AddOK(launch.PNameTimeSyncer, launch.PStartTimeSyncer, launch.PCloseTimeSyncer, launch.PNameDesign).
		AddOK(launch.PNameLocal, launch.PLocal, nil, launch.PNameDesign).
		AddOK(launch.PNameBlockItemReaders, launch.PBlockItemReaders, nil, launch.PNameDesign).
		AddOK(launch.PNameStorage, launch.PStorage, launch.PCloseStorage, launch.PNameLocal)
	_ = pps.POK(launch.PNameEncoder).PostAddOK(launch.PNameAddHinters, steps.PAddHinters)
	_ = pps.POK(launch.PNameDesign).
		PostAddOK(launch.PNameCheckDesign, launch.PCheckDesign).
		PostAddOK(launch.PNameINITObjectCache, launch.PINITObjectCache)
	_ = pps.POK(launch.PNameBlockItemReaders).
		PreAddOK(launch.PNameBlockItemReadersDecompressFunc, launch.PBlockItemReadersDecompressFunc).
		PostAddOK(launch.PNameRemotesBlockItemReaderFunc, launch.PRemotesBlockItemReaderFunc)
	_ = pps.POK(launch.PNameStorage).
		PreAddOK(launch.PNameCheckLocalFS, launch.PCheckAndCreateLocalFS).
		PreAddOK(launch.PNameLoadDatabase, launch.PLoadDatabase).
		PostAddOK(launch.PNameCheckLeveldbStorage, launch.PCheckLeveldbStorage).
		PostAddOK(launch.PNameLoadFromDatabase, launch.PLoadFromDatabase).
		PostAddOK(launch.PNameCheckBlocksOfStorage, launch.PCheckBlocksOfStorage).
		PostAddOK(launch.PNamePatchBlockItemReaders, launch.PPatchBlockItemReaders)
	return pps
}
