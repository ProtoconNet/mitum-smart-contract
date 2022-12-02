package cmds

import (
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/util/ps"
)

func DefaultINITPS() *ps.PS {
	pps := ps.NewPS("cmd-init")

	_ = pps.
		AddOK(launch.PNameEncoder, PEncoder, nil).
		AddOK(launch.PNameDesign, PLoadDesign, nil, launch.PNameEncoder).
		AddOK(launch.PNameTimeSyncer, PStartTimeSyncer /*launch.PCloseTimeSyncer, */, nil, launch.PNameDesign).
		AddOK(launch.PNameLocal, PLocal, nil, launch.PNameDesign).
		AddOK(launch.PNameStorage, launch.PStorage, launch.PCloseStorage, launch.PNameLocal).
		AddOK(PNameGenerateGenesis, PGenerateGenesis, nil, launch.PNameStorage)

	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, PAddHinters)

	_ = pps.POK(launch.PNameDesign).
		PostAddOK(launch.PNameCheckDesign, PCheckDesign).
		PostAddOK(launch.PNameGenesisDesign, launch.PGenesisDesign)

	_ = pps.POK(launch.PNameStorage).
		PreAddOK(launch.PNameCleanStorage, PCleanStorage).
		PreAddOK(launch.PNameCreateLocalFS, PCreateLocalFS).
		PreAddOK(launch.PNameLoadDatabase, PLoadDatabase)

	return pps
}
