package cmds

import (
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/util/ps"
)

func DefaultINITPS() *ps.PS {
	pps := ps.NewPS("cmd-init")

	_ = pps.
		AddOK(launch.PNameEncoder, PEncoder, nil).
		AddOK(launch.PNameDesign, launch.PLoadDesign, nil, launch.PNameEncoder).
		AddOK(PNameDigestDesign, PLoadDigestDesign, nil, launch.PNameEncoder).
		AddOK(launch.PNameTimeSyncer, launch.PStartTimeSyncer /*launch.PCloseTimeSyncer, */, nil, launch.PNameDesign).
		AddOK(launch.PNameLocal, launch.PLocal, nil, launch.PNameDesign).
		AddOK(launch.PNameStorage, launch.PStorage, launch.PCloseStorage, launch.PNameLocal).
		AddOK(PNameGenerateGenesis, PGenerateGenesis, nil, launch.PNameStorage)

	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, PAddHinters)

	_ = pps.POK(launch.PNameDesign).
		PostAddOK(launch.PNameCheckDesign, launch.PCheckDesign).
		PostAddOK(launch.PNameGenesisDesign, launch.PGenesisDesign)

	_ = pps.POK(launch.PNameStorage).
		PreAddOK(launch.PNameCleanStorage, launch.PCleanStorage).
		PreAddOK(launch.PNameCreateLocalFS, launch.PCreateLocalFS).
		PreAddOK(launch.PNameLoadDatabase, launch.PLoadDatabase)

	return pps
}
