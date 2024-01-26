package cmds

import (
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util/ps"
)

var (
	PNameDigest           = ps.Name("digest")
	PNameDigestStart      = ps.Name("digest_star")
	PNameMongoDBsDataBase = ps.Name("mongodb_database")
)

func DefaultRunPS() *ps.PS {
	pps := ps.NewPS("cmd-run")

	_ = pps.
		AddOK(launch.PNameEncoder, PEncoder, nil).
		AddOK(launch.PNameDesign, launch.PLoadDesign, nil, launch.PNameEncoder).
		AddOK(PNameDigestDesign, PLoadDigestDesign, nil, launch.PNameEncoder).
		AddOK(launch.PNameTimeSyncer, launch.PStartTimeSyncer, launch.PCloseTimeSyncer, launch.PNameDesign).
		AddOK(launch.PNameLocal, launch.PLocal, nil, launch.PNameDesign).
		AddOK(launch.PNameStorage, launch.PStorage, nil, launch.PNameLocal).
		AddOK(launch.PNameProposalMaker, PProposalMaker, nil, launch.PNameStorage).
		AddOK(launch.PNameNetwork, launch.PNetwork, nil, launch.PNameStorage).
		AddOK(launch.PNameMemberlist, launch.PMemberlist, nil, launch.PNameNetwork).
		AddOK(launch.PNameBlockItemReaders, launch.PBlockItemReaders, nil, launch.PNameDesign).
		AddOK(launch.PNameStartStorage, launch.PStartStorage, launch.PCloseStorage, launch.PNameStartNetwork).
		AddOK(launch.PNameStartNetwork, launch.PStartNetwork, launch.PCloseNetwork, launch.PNameStates).
		AddOK(launch.PNameStartMemberlist, launch.PStartMemberlist, launch.PCloseMemberlist, launch.PNameStartNetwork).
		AddOK(launch.PNameStartSyncSourceChecker, launch.PStartSyncSourceChecker, launch.PCloseSyncSourceChecker, launch.PNameStartNetwork).
		AddOK(launch.PNameStartLastConsensusNodesWatcher,
			launch.PStartLastConsensusNodesWatcher, launch.PCloseLastConsensusNodesWatcher, launch.PNameStartNetwork).
		AddOK(launch.PNameStates, launch.PStates, nil, launch.PNameNetwork).
		AddOK(launch.PNameStatesReady, nil, launch.PCloseStates,
			launch.PNameStartStorage,
			launch.PNameStartSyncSourceChecker,
			launch.PNameStartLastConsensusNodesWatcher,
			launch.PNameStartMemberlist,
			launch.PNameStartNetwork,
			launch.PNameStates).
		AddOK(PNameMongoDBsDataBase, ProcessDatabase, nil, PNameDigestDesign, launch.PNameStorage).
		AddOK(PNameDigester, ProcessDigester, nil, PNameMongoDBsDataBase).
		AddOK(PNameDigest, ProcessDigestAPI, nil, PNameDigestDesign, PNameMongoDBsDataBase, launch.PNameMemberlist).
		AddOK(PNameDigestStart, ProcessStartDigestAPI, nil, PNameDigest).
		AddOK(PNameStartDigester, ProcessStartDigester, nil, PNameDigestStart)

	_ = pps.POK(launch.PNameDesign).
		PostAddOK(launch.PNameCheckDesign, launch.PCheckDesign).
		PostAddOK(launch.PNameINITObjectCache, launch.PINITObjectCache)

	_ = pps.POK(launch.PNameLocal).
		PostAddOK(launch.PNameDiscoveryFlag, launch.PDiscoveryFlag).
		PostAddOK(launch.PNameLoadACL, launch.PLoadACL)

	_ = pps.POK(launch.PNameBlockItemReaders).
		PreAddOK(launch.PNameBlockItemReadersDecompressFunc, launch.PBlockItemReadersDecompressFunc).
		PostAddOK(launch.PNameRemotesBlockItemReaderFunc, launch.PRemotesBlockItemReaderFunc)

	_ = pps.POK(launch.PNameStorage).
		PreAddOK(launch.PNameCheckLocalFS, launch.PCheckAndCreateLocalFS).
		PreAddOK(launch.PNameLoadDatabase, launch.PLoadDatabase).
		PostAddOK(launch.PNameCheckLeveldbStorage, launch.PCheckLeveldbStorage).
		PostAddOK(launch.PNameLoadFromDatabase, launch.PLoadFromDatabase).
		PostAddOK(launch.PNameCheckBlocksOfStorage, launch.PCheckBlocksOfStorage).
		PostAddOK(launch.PNamePatchBlockItemReaders, launch.PPatchBlockItemReaders).
		PostAddOK(launch.PNameNodeInfo, launch.PNodeInfo)

	_ = pps.POK(launch.PNameNetwork).
		PreAddOK(launch.PNameQuicstreamClient, launch.PQuicstreamClient).
		PostAddOK(launch.PNameSyncSourceChecker, launch.PSyncSourceChecker).
		PostAddOK(launch.PNameSuffrageCandidateLimiterSet, launch.PSuffrageCandidateLimiterSet)

	_ = pps.POK(launch.PNameMemberlist).
		PreAddOK(launch.PNameLastConsensusNodesWatcher, launch.PLastConsensusNodesWatcher).
		PreAddOK(launch.PNameRateLimiterContextKey, launch.PNetworkRateLimiter).
		PostAddOK(launch.PNameBallotbox, launch.PBallotbox).
		PostAddOK(launch.PNameLongRunningMemberlistJoin, launch.PLongRunningMemberlistJoin).
		PostAddOK(launch.PNameSuffrageVoting, launch.PSuffrageVoting).
		PostAddOK(launch.PNameEventLoggingNetworkHandlers, launch.PEventLoggingNetworkHandlers)

	_ = pps.POK(launch.PNameStates).
		PreAddOK(launch.PNameProposerSelector, launch.PProposerSelector).
		PreAddOK(launch.PNameOperationProcessorsMap, POperationProcessorsMap).
		PreAddOK(launch.PNameNetworkHandlers, PNetworkHandlers).
		PreAddOK(launch.PNameNodeInConsensusNodesFunc, launch.PNodeInConsensusNodesFunc).
		PreAddOK(launch.PNameProposalProcessors, launch.PProposalProcessors).
		PreAddOK(launch.PNameBallotStuckResolver, launch.PBallotStuckResolver).
		PostAddOK(launch.PNamePatchLastConsensusNodesWatcher, launch.PPatchLastConsensusNodesWatcher).
		PostAddOK(launch.PNameStatesSetHandlers, launch.PStatesSetHandlers).
		PostAddOK(launch.PNameNetworkHandlersReadWriteNode, launch.PNetworkHandlersReadWriteNode).
		PostAddOK(launch.PNamePatchMemberlist, launch.PPatchMemberlist).
		PostAddOK(launch.PNameStatesNetworkHandlers, PStatesNetworkHandlers).
		PostAddOK(launch.PNameHandoverNetworkHandlers, launch.PHandoverNetworkHandlers)

	return pps
}
