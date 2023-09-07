package digest

import (
	"context"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/network/quicstream"
	quicstreamheader "github.com/ProtoconNet/mitum2/network/quicstream/header"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
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
		hd.Log().Err(err).Msg("get node info")

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

	var nodeInfoList []map[string]interface{}
	switch {
	case err != nil:
		return nil, err

	default:
		var nodeList []quicstream.ConnInfo
		memberList.Members(func(node quicmemberlist.Member) bool {
			nodeList = append(nodeList, node.ConnInfo())
			return true
		})
		for i := range nodeList {
			nodeInfo, err := NodeInfo(client, nodeList[i])
			if err != nil {
				return nil, err
			}
			nodeInfoList = append(nodeInfoList, *nodeInfo)
		}
	}

	if i, err := hd.buildNodeInfoHal(nodeInfoList); err != nil {
		return nil, err
	} else {
		return hd.enc.Marshal(i)
	}
}

func (hd *Handlers) buildNodeInfoHal(ni []map[string]interface{}) (Hal, error) {
	var hal Hal = NewBaseHal(ni, NewHalLink(HandlerPathNodeInfo, nil))

	return hal, nil
}

func NodeInfo(client *isaacnetwork.BaseClient, connInfo quicstream.ConnInfo) (*map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*9)
	defer cancel()

	stream, _, err := client.Dial(ctx, connInfo)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = client.Close()
	}()

	header := isaacnetwork.NewNodeInfoRequestHeader()

	var nodeInfo *map[string]interface{}
	err = stream(ctx, func(ctx context.Context, broker *quicstreamheader.ClientBroker) error {
		if err := broker.WriteRequestHead(ctx, header); err != nil {
			return err
		}

		var enc encoder.Encoder

		switch rEnc, rh, err := broker.ReadResponseHead(ctx); {
		case err != nil:
			return err
		case !rh.OK():
			return errors.Errorf("not ok")
		case rh.Err() != nil:
			return rh.Err()
		default:
			enc = rEnc
		}

		switch bodyType, bodyLength, r, err := broker.ReadBodyErr(ctx); {
		case err != nil:
			return err
		case bodyType == quicstreamheader.EmptyBodyType,
			bodyType == quicstreamheader.FixedLengthBodyType && bodyLength < 1:
			return errors.Errorf("empty body")
		default:
			var v interface{}
			if err := enc.StreamDecoder(r).Decode(&v); err != nil {
				return err
			}

			ni, ok := v.(map[string]interface{})
			if !ok {
				return errors.Errorf("expected map[string]interface{}, not %T", v)
			}

			nodeInfo = &ni

			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	return nodeInfo, nil
}
