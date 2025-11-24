package digest

import (
	"net/http"
	"time"

	ptypes "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
)

func (hd *Handlers) handleContractDesign(w http.ResponseWriter, r *http.Request) {
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	contract, err, status := ParseRequest(w, r, "contract")
	if err != nil {
		HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handleContractDesignInGroup(contract)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cacheKey, time.Second*3)
		}
	}
}

func (hd *Handlers) handleContractDesignInGroup(contract string) ([]byte, error) {
	var de ptypes.Design
	var st base.State

	de, st, err := ContractDesign(hd.database, contract)
	if err != nil {
		return nil, err
	}

	i, err := hd.buildContractDesign(contract, de, st)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(i)
}

func (hd *Handlers) buildContractDesign(contract string, de ptypes.Design, st base.State) (Hal, error) {
	h, err := hd.combineURL(HandlerPathContractDesign, "contract", contract)
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(de, NewHalLink(h, nil))

	h, err = hd.combineURL(HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.combineURL(HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", NewHalLink(h, nil))
	}

	return hal, nil
}

func (hd *Handlers) handleContractData(w http.ResponseWriter, r *http.Request) {
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	contract, err, status := ParseRequest(w, r, "contract")
	if err != nil {
		HTTP2ProblemWithError(w, err, status)
		return
	}

	dataKey, err, status := ParseRequest(w, r, "data_key")
	if err != nil {
		HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handleContractDataInGroup(contract, dataKey)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cacheKey, time.Second*3)
		}
	}
}

func (hd *Handlers) handleContractDataInGroup(contract, key string) ([]byte, error) {
	data, st, err := ContractData(hd.database, contract, key)
	if err != nil {
		return nil, err
	}

	i, err := hd.buildContractData(contract, data, st, key)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(i)
}

func (hd *Handlers) buildContractData(
	contract string, data map[string]interface{}, st base.State, key string) (Hal, error) {
	h, err := hd.combineURL(
		HandlerPathContractData,
		"contract", contract, "data_key", key)
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(data, NewHalLink(h, nil))
	h, err = hd.combineURL(HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.combineURL(HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", NewHalLink(h, nil))
	}

	return hal, nil
}
