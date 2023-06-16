package digest

import (
	"context"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/network/quicstream"
	"net/http"
	"time"

	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
)

func (hd *Handlers) SetNodeInfoHandler(handler NodeInfoHandler) *Handlers {
	hd.nodeInfoHandler = handler

	return hd
}

func (hd *Handlers) handleNodeInfo(w http.ResponseWriter, r *http.Request) {

	//if hd.nodeInfoHandler == nil {
	//	HTTP2NotSupported(w, nil)
	//
	//	return
	//}

	cachekey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cachekey, hd.handleNodeInfoInGroup); err != nil {
		hd.Log().Err(err).Msg("failed to get node info")

		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cachekey, time.Second*3)
		}
	}
}

func (hd *Handlers) handleNodeInfoInGroup() (interface{}, error) {
	client, memberList, err := hd.client()

	var nodeInfoList []isaacnetwork.NodeInfo
	switch {
	case err != nil:
		return nil, err

	default:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		var nodeList []quicstream.UDPConnInfo
		memberList.Members(func(node quicmemberlist.Member) bool {
			nodeList = append(nodeList, node.UDPConnInfo())
			return true
		})
		for i := range nodeList {
			nodeInfo, _, err := client.NodeInfo(ctx, nodeList[i])
			if err != nil {
				return nil, err
			}
			nodeInfoList = append(nodeInfoList, nodeInfo)
		}
	}

	if i, err := hd.buildNodeInfoHal(nodeInfoList); err != nil {
		return nil, err
	} else {
		return hd.enc.Marshal(i)
	}
}

func (hd *Handlers) buildNodeInfoHal(ni []isaacnetwork.NodeInfo) (Hal, error) {
	var hal Hal = NewBaseHal(ni, NewHalLink(HandlerPathNodeInfo, nil))

	return hal, nil
}
