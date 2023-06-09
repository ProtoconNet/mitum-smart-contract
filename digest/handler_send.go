package digest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

func (hd *Handlers) SetSend(f func(interface{}) (base.Operation, error)) *Handlers {
	hd.send = f

	return hd
}

func (hd *Handlers) handleSend(w http.ResponseWriter, r *http.Request) {
	if hd.send == nil {
		HTTP2NotSupported(w, nil)

		return
	}

	body := &bytes.Buffer{}
	if _, err := io.Copy(body, r.Body); err != nil {
		HTTP2ProblemWithError(w, err, http.StatusInternalServerError)
		return
	}

	var hal Hal
	var v json.RawMessage
	if err := Unmarshal(body.Bytes(), &v); err != nil {
		HTTP2ProblemWithError(w, err, http.StatusBadRequest)
		return
	} else if hinter, err := hd.enc.Decode(body.Bytes()); err != nil {
		HTTP2ProblemWithError(w, err, http.StatusBadRequest)
		return
	} else if h, err := hd.sendItem(hinter); err != nil {
		HTTP2ProblemWithError(w, err, http.StatusBadRequest)
		return
	} else {
		hal = h
	}
	HTTP2WriteHal(hd.enc, w, hal, http.StatusOK)
}

func (hd *Handlers) sendItem(v interface{}) (Hal, error) {
	switch t := v.(type) {
	case base.Operation:
		if err := t.IsValid(hd.networkID); err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("unsupported message type, %T", v)
	}

	return hd.sendOperation(v)
}

func (hd *Handlers) sendOperation(v interface{}) (Hal, error) {
	op, err := hd.send(v)
	if err != nil {
		return nil, err
	}
	return hd.buildSealHal(op)
}

func (hd *Handlers) buildSealHal(op base.Operation) (Hal, error) {
	var hal Hal = NewBaseHal(op, HalLink{})
	/*
		if t, ok := sl.(operation.Seal); ok {
			for i := range t.Operations() {
				op := t.Operations()[i]
				h, err := hd.combineURL(HandlerPathOperation, "hash", op.Fact().Hash().String())
				if err != nil {
					return nil, err
				}
				hal.AddLink(fmt.Sprintf("operation:%d", i), NewHalLink(h, nil))
			}
		}
	*/

	return hal, nil
}
