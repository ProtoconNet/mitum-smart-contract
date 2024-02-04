package digest

import (
	"fmt"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"net/http"
	"strings"
	"time"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (hd *Handlers) handleCurrencies(w http.ResponseWriter, r *http.Request) {
	cachekey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		return hd.handleCurrenciesInGroup()
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cachekey, time.Second*3)
		}
	}
}

func (hd *Handlers) handleCurrenciesInGroup() ([]byte, error) {
	var hal Hal = NewBaseHal(nil, NewHalLink(HandlerPathCurrencies, nil))
	hal = hal.AddLink("currency:{currencyid}", NewHalLink(HandlerPathCurrency, nil).SetTemplated())

	cids, err := hd.database.currencies()
	if err != nil {
		return nil, err
	}
	for i := range cids {
		h, err := hd.combineURL(HandlerPathCurrency, "currencyid", cids[i])
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink(fmt.Sprintf("currency:%s", cids[i]), NewHalLink(h, nil))
	}

	return hd.enc.Marshal(hal)
}

func (hd *Handlers) handleCurrency(w http.ResponseWriter, r *http.Request) {
	cachekey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	var cid string
	s, found := mux.Vars(r)["currencyid"]
	if !found {
		HTTP2ProblemWithError(w, errors.Errorf("empty currency id"), http.StatusBadRequest)

		return
	}

	s = strings.TrimSpace(s)
	if len(s) < 1 {
		HTTP2ProblemWithError(w, errors.Errorf("empty currency id"), http.StatusBadRequest)

		return
	}
	cid = s

	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		return hd.handleCurrencyInGroup(cid)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cachekey, time.Second*3)
		}
	}
}

func (hd *Handlers) handleCurrencyInGroup(cid string) ([]byte, error) {
	var de types.CurrencyDesign
	var st base.State

	de, st, err := hd.database.currency(cid)
	if err != nil {
		return nil, err
	}

	i, err := hd.buildCurrency(de, st)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(i)
}

func (hd *Handlers) buildCurrency(de types.CurrencyDesign, st base.State) (Hal, error) {
	h, err := hd.combineURL(HandlerPathCurrency, "currencyid", de.Currency().String())
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(de, NewHalLink(h, nil))

	hal = hal.AddLink("currency:{currencyid}", NewHalLink(HandlerPathCurrency, nil).SetTemplated())

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
