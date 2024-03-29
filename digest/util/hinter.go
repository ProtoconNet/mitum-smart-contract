package util

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

type BaseHinter struct {
	HT hint.Hint `bson:"_hint"` //nolint:tagliatelle //...
}

func NewBaseHinter(ht hint.Hint) BaseHinter {
	return BaseHinter{HT: ht}
}

func (ht BaseHinter) Hint() hint.Hint {
	return ht.HT
}

func (BaseHinter) SetHint(n hint.Hint) hint.Hinter {
	return BaseHinter{HT: n}
}

func (ht BaseHinter) IsValid(expectedType []byte) error {
	if err := ht.HT.IsValid(nil); err != nil {
		return errors.WithMessage(err, "invalid hint in BaseHinter")
	}

	if len(expectedType) > 0 {
		if t := hint.Type(string(expectedType)); t != ht.HT.Type() {
			return util.ErrInvalid.Errorf("Type does not match in BaseHinter, %q != %q", ht.HT.Type(), t)
		}
	}

	return nil
}

func (ht BaseHinter) Bytes() []byte {
	return ht.HT.Bytes()
}
