package digest

import (
	"net/http"
	"time"

	capi "github.com/imfact-labs/currency-model/api"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/types"
)

func (hd *Handlers) handleContractDesign(w http.ResponseWriter, r *http.Request) {
	cacheKey := capi.CacheKeyPath(r)
	if err := capi.LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	contract, err, status := capi.ParseRequest(w, r, "contract")
	if err != nil {
		capi.HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handleContractDesignInGroup(contract)
	}); err != nil {
		capi.HTTP2HandleError(w, err)
	} else {
		capi.HTTP2WriteHalBytes(hd.encoder, w, v.([]byte), http.StatusOK)

		if !shared {
			capi.HTTP2WriteCache(w, cacheKey, time.Second*3)
		}
	}
}

func (hd *Handlers) handleContractDesignInGroup(contract string) ([]byte, error) {
	var de types.Design
	var st base.State

	de, st, err := ContractDesign(hd.database, contract)
	if err != nil {
		return nil, err
	}

	i, err := hd.buildContractDesign(contract, de, st)
	if err != nil {
		return nil, err
	}
	return hd.encoder.Marshal(i)
}

func (hd *Handlers) buildContractDesign(contract string, de types.Design, st base.State) (capi.Hal, error) {
	h, err := hd.combineURL(HandlerPathContractDesign, "contract", contract)
	if err != nil {
		return nil, err
	}

	var hal capi.Hal
	hal = capi.NewBaseHal(de, capi.NewHalLink(h, nil))

	h, err = hd.combineURL(capi.HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", capi.NewHalLink(h, nil))

	h, err = hd.combineURL(HandlerPathContractQuery, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("query", capi.NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.combineURL(capi.HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", capi.NewHalLink(h, nil))
	}

	return hal, nil
}
