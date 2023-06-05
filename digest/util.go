package digest

import (
	"fmt"
	base3 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ProtoconNet/mitum2/base"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var JSON = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            false,
	ValidateJsonRawMessage: false,
}.Froze()

func Marshal(v interface{}) ([]byte, error) {
	return JSON.Marshal(v)
}

func Unmarshal(b []byte, v interface{}) error {
	return JSON.Unmarshal(b, v)
}

func IsAccountState(st base.State) (base3.Account, bool, error) {
	if !currency.IsStateAccountKey(st.Key()) {
		return base3.Account{}, false, nil
	}

	ac, err := currency.LoadStateAccountValue(st)
	if err != nil {
		return base3.Account{}, false, err
	}
	return ac, true, nil
}

func IsBalanceState(st base.State) (base3.Amount, bool, error) {
	if !currency.IsStateBalanceKey(st.Key()) {
		return base3.Amount{}, false, nil
	}

	am, err := currency.StateBalanceValue(st)
	if err != nil {
		return base3.Amount{}, false, err
	}
	return am, true, nil
}

func parseHeightFromPath(s string) (base.Height, error) {
	s = strings.TrimSpace(s)

	if len(s) < 1 {
		return base.NilHeight, errors.Errorf("empty height")
	} else if len(s) > 1 && strings.HasPrefix(s, "0") {
		return base.NilHeight, errors.Errorf("invalid height, %q", s)
	}

	return base.ParseHeightString(s)
}

func parseHashFromPath(s string) (util.Hash, error) {
	s = strings.TrimSpace(s)
	if len(s) < 1 {
		return nil, errors.Errorf("empty hash")
	}

	h := valuehash.NewBytesFromString(s)
	if err := h.IsValid(nil); err != nil {
		return nil, err
	}

	return h, nil
}

func parseLimitQuery(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return int64(-1)
	}
	return n
}

func parseStringQuery(s string) string {
	return strings.TrimSpace(s)
}

func stringOffsetQuery(offset string) string {
	return fmt.Sprintf("offset=%s", offset)
}

func parseBoolQuery(s string) bool {
	return s == "1"
}

func stringBoolQuery(key string, v bool) string { // nolint:unparam
	if v {
		return fmt.Sprintf("%s=1", key)
	}

	return ""
}

func addQueryValue(b, s string) string {
	if len(s) < 1 {
		return b
	}

	if !strings.Contains(b, "?") {
		return b + "?" + s
	}

	return b + "&" + s
}

func HTTP2Stream(enc encoder.Encoder, w http.ResponseWriter, bufsize int, status int) (*jsoniter.Stream, func()) {
	w.Header().Set(HTTP2EncoderHintHeader, enc.Hint().String())
	w.Header().Set("Content-Type", HALMimetype)

	if status != http.StatusOK {
		w.WriteHeader(status)
	}

	stream := jsoniter.NewStream(HALJSONConfigDefault, w, bufsize)
	return stream, func() {
		_ = stream.Flush()
	}
}

func HTTP2NotSupported(w http.ResponseWriter, err error) {
	if err == nil {
		err = util.NewError("not supported")
	}

	HTTP2ProblemWithError(w, err, http.StatusInternalServerError)
}

func HTTP2ProblemWithError(w http.ResponseWriter, err error, status int) {
	HTTP2WritePoblem(w, NewProblemFromError(err), status)
}

func HTTP2WritePoblem(w http.ResponseWriter, pr Problem, status int) {
	if status == 0 {
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", ProblemMimetype)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	var output []byte
	if b, err := Marshal(pr.title); err != nil {
		output = unknownProblemJSON
	} else {
		output = b
	}

	w.WriteHeader(status)
	_, _ = w.Write(output)
}

func HTTP2WriteHal(enc encoder.Encoder, w http.ResponseWriter, hal Hal, status int) { // nolint:unparam
	stream, flush := HTTP2Stream(enc, w, 1, status)
	defer flush()

	stream.WriteVal(hal)
}

func HTTP2WriteHalBytes(enc encoder.Encoder, w http.ResponseWriter, b []byte, status int) { // nolint:unparam
	w.Header().Set(HTTP2EncoderHintHeader, enc.Hint().String())
	w.Header().Set("Content-Type", HALMimetype)

	if status != http.StatusOK {
		w.WriteHeader(status)
	}

	_, _ = w.Write(b)
}

func HTTP2WriteCache(w http.ResponseWriter, key string, expire time.Duration) {
	if expire < 1 {
		return
	}

	if cw, ok := w.(*CacheResponseWriter); ok {
		_ = cw.SetKey(key).SetExpire(expire)
	}
}

func HTTP2HandleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, util.NewError("not found")):
		status = http.StatusNotFound
	case errors.Is(err, util.NewError("bad request")):
		status = http.StatusBadRequest
	case errors.Is(err, util.NewError("not supported")):
		status = http.StatusInternalServerError
	}

	HTTP2ProblemWithError(w, err, status)
}

type NodeInfoHandler func() (isaacnetwork.NodeInfo, error)
