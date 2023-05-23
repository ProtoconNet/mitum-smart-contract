package base

import "github.com/ProtoconNet/mitum2/util"

func (a Big) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(a.String())
}

func (a *Big) UnmarshalJSON(b []byte) error {
	var s string
	if err := util.UnmarshalJSON(b, &s); err != nil {
		return err
	}

	i, err := NewBigFromString(s)
	if err != nil {
		return err
	}
	*a = i

	return nil
}
