package digest

import (
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
)

func (hd *Handlers) SetNetworkClientFunc(f func() (*isaacnetwork.QuicstreamClient, *quicmemberlist.Memberlist, error)) *Handlers {
	hd.client = f
	return hd
}
