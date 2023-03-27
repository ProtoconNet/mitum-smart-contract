package digest

import (
	"time"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	ManifestValueHint = hint.MustNewHint("mitum-currency-manifest-value-v0.0.1")
)

type ManifestValue struct {
	hint.BaseHinter
	manifest    base.Manifest
	height      base.Height
	confirmedAt time.Time
}

func NewManifestValue(
	manifest base.Manifest,
	height base.Height,
	confirmedAt time.Time,
) ManifestValue {
	return ManifestValue{
		BaseHinter:  hint.NewBaseHinter(ManifestValueHint),
		manifest:    manifest,
		height:      height,
		confirmedAt: confirmedAt,
	}
}

func (ManifestValue) Hint() hint.Hint {
	return OperationValueHint
}

func (va ManifestValue) Manifest() base.Manifest {
	return va.manifest
}

func (va ManifestValue) Height() base.Height {
	return va.height
}

func (va ManifestValue) ConfirmedAt() time.Time {
	return va.confirmedAt
}
