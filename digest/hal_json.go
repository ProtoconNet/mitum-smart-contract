package digest

import (
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
)

var HALJSONConfigDefault = jsoniter.Config{
	EscapeHTML: false,
}.Froze()

type BaseHalJSONMarshaler struct {
	hint.BaseHinter
	I  interface{}            `json:"_embedded,omitempty"`
	LS map[string]HalLink     `json:"_links,omitempty"`
	EX map[string]interface{} `json:"_extra,omitempty"`
}

func (hal BaseHal) MarshalJSON() ([]byte, error) {
	ls := hal.Links()
	ls["self"] = hal.Self()

	return util.MarshalJSON(BaseHalJSONMarshaler{
		BaseHinter: hal.BaseHinter,
		I:          hal.i,
		LS:         ls,
		EX:         hal.extras,
	})
}

type BaseHalJSONUnpacker struct {
	R  json.RawMessage        `json:"_embedded,omitempty"`
	LS map[string]HalLink     `json:"_links,omitempty"`
	EX map[string]interface{} `json:"_extra,omitempty"`
}

func (hal *BaseHal) UnmarshalJSON(b []byte) error {
	var uh BaseHalJSONUnpacker
	if err := Unmarshal(b, &uh); err != nil {
		return err
	}

	hal.raw = uh.R
	hal.links = uh.LS
	hal.extras = uh.EX

	return nil
}

func (hl HalLink) MarshalJSON() ([]byte, error) {
	all := map[string]interface{}{}
	if hl.properties != nil {
		for k := range hl.properties {
			all[k] = hl.properties[k]
		}
	}
	all["href"] = hl.href

	return Marshal(all)
}

type HalLinkJSONUnpacker struct {
	HR string                 `json:"href"`
	PR map[string]interface{} `json:"properties,omitempty"`
}

func (hl *HalLink) UnmarshalJSON(b []byte) error {
	var uh HalLinkJSONUnpacker
	if err := Unmarshal(b, &uh); err != nil {
		return err
	}

	hl.href = uh.HR
	hl.properties = uh.PR

	return nil
}
