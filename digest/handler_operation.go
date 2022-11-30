package digest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	mitumutil "github.com/spikeekips/mitum/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (hd *Handlers) handleOperation(w http.ResponseWriter, r *http.Request) {
	fmt.Println(">>>>>>> handleOperation 1")
	cachekey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}
	fmt.Println(">>>>>>> handleOperation 2")
	h, err := parseHashFromPath(mux.Vars(r)["hash"])
	if err != nil {
		HTTP2ProblemWithError(w, errors.Wrap(err, "invalid hash for operation by hash"), http.StatusBadRequest)

		return
	}
	fmt.Println(">>>>>>> handleOperation 3")
	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		return hd.handleOperationInGroup(h)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cachekey, time.Hour*30)
		}
	}
}

func (hd *Handlers) handleOperationInGroup(h mitumutil.Hash) ([]byte, error) {
	fmt.Println(">>>>>>> handleOperationInGroup 1")
	switch va, found, err := hd.database.Operation(h, true); {
	case err != nil:
		fmt.Println(">>>>>>> handleOperationInGroup 2", err)
		return nil, err
	case !found:
		fmt.Println(">>>>>>> handleOperationInGroup 3")
		return nil, mitumutil.ErrNotFound.Errorf("operation not found")
	default:
		fmt.Println(">>>>>>> handleOperationInGroup 4")
		hal, err := hd.buildOperationHal(va)
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operation:{hash}", NewHalLink(HandlerPathOperation, nil).SetTemplated())
		hal = hal.AddLink("block:{height}", NewHalLink(HandlerPathBlockByHeight, nil).SetTemplated())
		fmt.Println(">>>>>>> handleOperationInGroup 3", hal)
		return hd.enc.Marshal(hal)
	}
}

func (hd *Handlers) handleOperations(w http.ResponseWriter, r *http.Request) {
	offset := parseOffsetQuery(r.URL.Query().Get("offset"))
	reverse := parseBoolQuery(r.URL.Query().Get("reverse"))

	cachekey := CacheKey(r.URL.Path, stringOffsetQuery(offset), stringBoolQuery("reverse", reverse))
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		i, filled, err := hd.handleOperationsInGroup(offset, reverse)

		return []interface{}{i, filled}, err
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		var b []byte
		var filled bool
		{
			l := v.([]interface{})
			b = l[0].([]byte)
			filled = l[1].(bool)
		}

		HTTP2WriteHalBytes(hd.enc, w, b, http.StatusOK)

		if !shared {
			expire := hd.expireNotFilled
			if len(offset) > 0 && filled {
				expire = time.Hour * 30
			}

			HTTP2WriteCache(w, cachekey, expire)
		}
	}
}

func (hd *Handlers) handleOperationsInGroup(offset string, reverse bool) ([]byte, bool, error) {
	filter, err := buildOperationsFilterByOffset(offset, reverse)
	if err != nil {
		return nil, false, err
	}

	var vas []Hal
	switch l, e := hd.loadOperationsHALFromDatabase(filter, reverse); {
	case e != nil:
		return nil, false, e
	case len(l) < 1:
		return nil, false, mitumutil.ErrNotFound.Errorf("operations not found")
	default:
		vas = l
	}

	h, err := hd.combineURL(HandlerPathOperations)
	if err != nil {
		return nil, false, err
	}
	hal := hd.buildOperationsHal(h, vas, offset, reverse)
	if next := nextOffsetOfOperations(h, vas, reverse); len(next) > 0 {
		hal = hal.AddLink("next", NewHalLink(next, nil))
	}

	b, err := hd.enc.Marshal(hal)
	return b, int64(len(vas)) == hd.itemsLimiter("operations"), err
}

func (hd *Handlers) handleOperationsByHeight(w http.ResponseWriter, r *http.Request) {
	offset := parseOffsetQuery(r.URL.Query().Get("offset"))
	reverse := parseBoolQuery(r.URL.Query().Get("reverse"))

	cachekey := CacheKey(r.URL.Path, stringOffsetQuery(offset), stringBoolQuery("reverse", reverse))
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	var height base.Height
	switch h, err := parseHeightFromPath(mux.Vars(r)["height"]); {
	case err != nil:
		HTTP2ProblemWithError(w, errors.Errorf("invalid height found for manifest by height"), http.StatusBadRequest)

		return
	case h <= base.NilHeight:
		HTTP2ProblemWithError(w, errors.Errorf("invalid height, %v", h), http.StatusBadRequest)
		return
	default:
		height = h
	}

	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		i, filled, err := hd.handleOperationsByHeightInGroup(height, offset, reverse)
		return []interface{}{i, filled}, err
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		var b []byte
		var filled bool
		{
			l := v.([]interface{})
			b = l[0].([]byte)
			filled = l[1].(bool)
		}

		HTTP2WriteHalBytes(hd.enc, w, b, http.StatusOK)

		if !shared {
			expire := hd.expireNotFilled
			if len(offset) > 0 && filled {
				expire = time.Hour * 30
			}

			HTTP2WriteCache(w, cachekey, expire)
		}
	}
}

func (hd *Handlers) handleOperationsByHeightInGroup(
	height base.Height,
	offset string,
	reverse bool,
) ([]byte, bool, error) {
	filter, err := buildOperationsByHeightFilterByOffset(height, offset, reverse)
	if err != nil {
		return nil, false, err
	}

	var vas []Hal
	switch l, e := hd.loadOperationsHALFromDatabase(filter, reverse); {
	case e != nil:
		return nil, false, e
	case len(l) < 1:
		return nil, false, mitumutil.ErrNotFound.Errorf("operations not found")
	default:
		vas = l
	}

	h, err := hd.combineURL(HandlerPathOperationsByHeight, "height", height.String())
	if err != nil {
		return nil, false, err
	}
	hal := hd.buildOperationsHal(h, vas, offset, reverse)
	if next := nextOffsetOfOperationsByHeight(h, vas, reverse); len(next) > 0 {
		hal = hal.AddLink("next", NewHalLink(next, nil))
	}

	b, err := hd.enc.Marshal(hal)
	return b, int64(len(vas)) == hd.itemsLimiter("operations"), err
}

func (hd *Handlers) buildOperationHal(va OperationValue) (Hal, error) {
	var hal Hal
	fmt.Println(">>>>>> buildOperationHal 1", va)
	h, err := hd.combineURL(HandlerPathOperation, "hash", va.Operation().Fact().Hash().String())
	if err != nil {
		return nil, err
	}
	hal = NewBaseHal(va, NewHalLink(h, nil))
	fmt.Println(">>>>>> buildOperationHal 2")
	h, err = hd.combineURL(HandlerPathBlockByHeight, "height", va.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", NewHalLink(h, nil))
	fmt.Println(">>>>>> buildOperationHal 3", va.Height().String())
	h, err = hd.combineURL(HandlerPathManifestByHeight, "height", va.Height().String())
	if err != nil {
		return nil, err
	}
	fmt.Println(">>>>>> buildOperationHal 4")
	hal = hal.AddLink("manifest", NewHalLink(h, nil))
	fmt.Println(">>>>>> buildOperationHal 5")
	if va.InState() {
		if t, ok := va.Operation().(currency.CreateAccounts); ok {
			items := t.Fact().(currency.CreateAccountsFact).Items()
			for i := range items {
				a, err := items[i].Address()
				if err != nil {
					return nil, err
				}
				address := a.String()

				h, err := hd.combineURL(HandlerPathAccount, "address", address)
				if err != nil {
					return nil, err
				}
				keyHash := items[i].Keys().Hash().String()
				hal = hal.AddLink(
					fmt.Sprintf("new_account:%s", keyHash),
					NewHalLink(h, nil).
						SetProperty("key", keyHash).
						SetProperty("address", address),
				)
			}
		}
	}

	return hal, nil
}

func (*Handlers) buildOperationsHal(baseSelf string, vas []Hal, offset string, reverse bool) Hal {
	var hal Hal

	self := baseSelf
	if len(offset) > 0 {
		self = addQueryValue(baseSelf, stringOffsetQuery(offset))
	}
	if reverse {
		self = addQueryValue(self, stringBoolQuery("reverse", reverse))
	}
	hal = NewBaseHal(vas, NewHalLink(self, nil))

	hal = hal.AddLink("reverse", NewHalLink(addQueryValue(baseSelf, stringBoolQuery("reverse", !reverse)), nil))

	return hal
}

func buildOperationsFilterByOffset(offset string, reverse bool) (bson.M, error) {
	filter := bson.M{}
	if len(offset) > 0 {
		height, index, err := parseOffset(offset)
		if err != nil {
			return nil, err
		}

		if reverse {
			filter["$or"] = []bson.M{
				{"height": bson.M{"$lt": height}},
				{"$and": []bson.M{
					{"height": height},
					{"index": bson.M{"$lt": index}},
				}},
			}
		} else {
			filter["$or"] = []bson.M{
				{"height": bson.M{"$gt": height}},
				{"$and": []bson.M{
					{"height": height},
					{"index": bson.M{"$gt": index}},
				}},
			}
		}
	}

	return filter, nil
}

func buildOperationsByHeightFilterByOffset(height base.Height, offset string, reverse bool) (bson.M, error) {
	var filter bson.M
	if len(offset) < 1 {
		return bson.M{"height": height}, nil
	}

	index, err := strconv.ParseUint(offset, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid index of offset")
	}

	if reverse {
		filter = bson.M{
			"height": height,
			"index":  bson.M{"$lt": index},
		}
	} else {
		filter = bson.M{
			"height": height,
			"index":  bson.M{"$gt": index},
		}
	}

	return filter, nil
}

func nextOffsetOfOperations(baseSelf string, vas []Hal, reverse bool) string {
	var nextoffset string
	if len(vas) > 0 {
		va := vas[len(vas)-1].Interface().(OperationValue)
		nextoffset = buildOffset(va.Height(), va.Index())
	}

	if len(nextoffset) < 1 {
		return ""
	}

	next := baseSelf
	if len(nextoffset) > 0 {
		next = addQueryValue(next, stringOffsetQuery(nextoffset))
	}

	if reverse {
		next = addQueryValue(next, stringBoolQuery("reverse", reverse))
	}

	return next
}

func nextOffsetOfOperationsByHeight(baseSelf string, vas []Hal, reverse bool) string {
	var nextoffset string
	if len(vas) > 0 {
		va := vas[len(vas)-1].Interface().(OperationValue)
		nextoffset = fmt.Sprintf("%d", va.Index())
	}

	if len(nextoffset) < 1 {
		return ""
	}

	next := baseSelf
	if len(nextoffset) > 0 {
		next = addQueryValue(next, stringOffsetQuery(nextoffset))
	}

	if reverse {
		next = addQueryValue(next, stringBoolQuery("reverse", reverse))
	}

	return next
}

func (hd *Handlers) loadOperationsHALFromDatabase(filter bson.M, reverse bool) ([]Hal, error) {
	var vas []Hal
	if err := hd.database.Operations(
		filter, true, reverse, hd.itemsLimiter("operations"),
		func(_ mitumutil.Hash, va OperationValue) (bool, error) {
			hal, err := hd.buildOperationHal(va)
			if err != nil {
				return false, err
			}
			vas = append(vas, hal)

			return true, nil
		},
	); err != nil {
		return nil, err
	} else if len(vas) < 1 {
		return nil, nil
	}

	return vas, nil
}
