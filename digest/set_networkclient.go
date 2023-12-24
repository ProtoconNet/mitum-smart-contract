package digest

import (
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/network/quicstream"
)

//func (hd *Handlers) SetNetworkClientFunc(f func() (*isaacnetwork.BaseClient, *quicmemberlist.Memberlist, error)) *Handlers {
//	hd.client = f
//	return hd
//}

func (hd *Handlers) SetNetworkClientFunc(f func() (*isaacnetwork.BaseClient, *quicmemberlist.Memberlist, []quicstream.ConnInfo, error)) *Handlers {
	hd.client = f
	return hd
}
