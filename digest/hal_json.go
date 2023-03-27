package digest

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	jsoniter "github.com/json-iterator/go"
)

var HALJSONConfigDefault = jsoniter.Config{
	EscapeHTML: false,
}.Froze()

type BaseHalJSONMarshaler struct {
	hint.BaseHinter
	Embedded interface{}            `json:"_embedded,omitempty"`
	Links    map[string]HalLink     `json:"_links,omitempty"`
	Extra    map[string]interface{} `json:"_extra,omitempty"`
}

func (hal BaseHal) MarshalJSON() ([]byte, error) {
	ls := hal.Links()
	ls["self"] = hal.Self()

	return util.MarshalJSON(BaseHalJSONMarshaler{
		BaseHinter: hal.BaseHinter,
		Embedded:   hal.i,
		Links:      ls,
		Extra:      hal.extras,
	})
}

type BaseHalJSONUnpacker struct {
	Embedded json.RawMessage        `json:"_embedded,omitempty"`
	Links    map[string]HalLink     `json:"_links,omitempty"`
	Extra    map[string]interface{} `json:"_extra,omitempty"`
}

func (hal *BaseHal) UnmarshalJSON(b []byte) error {
	var uh BaseHalJSONUnpacker
	if err := Unmarshal(b, &uh); err != nil {
		return err
	}

	hal.raw = uh.Embedded
	hal.links = uh.Links
	hal.extras = uh.Extra

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
	Href       string                 `json:"href"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

func (hl *HalLink) UnmarshalJSON(b []byte) error {
	var uh HalLinkJSONUnpacker
	if err := Unmarshal(b, &uh); err != nil {
		return err
	}

	hl.href = uh.Href
	hl.properties = uh.Properties

	return nil
}
