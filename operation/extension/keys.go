package extension

import (
	"bytes"
	"sort"

	"github.com/ProtoconNet/mitum-currency/v3/base"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var ContractAccountKeysHint = hint.MustNewHint("mitum-currency-contract-account-keys-v0.0.1")

type ContractAccountKeys struct {
	hint.BaseHinter
	h         util.Hash
	keys      []base.AccountKey
	threshold uint
}

func EmptyBaseAccountKeys() ContractAccountKeys {
	return ContractAccountKeys{BaseHinter: hint.NewBaseHinter(ContractAccountKeysHint)}
}

func NewContractAccountKeys() (ContractAccountKeys, error) {
	ks := ContractAccountKeys{BaseHinter: hint.NewBaseHinter(ContractAccountKeysHint), keys: []base.AccountKey{}, threshold: 100}

	h, err := ks.GenerateHash()
	if err != nil {
		return ContractAccountKeys{}, err
	}
	ks.h = h

	return ks, ks.IsValid(nil)
}

func (ks ContractAccountKeys) Hash() util.Hash {
	return ks.h
}

func (ks ContractAccountKeys) GenerateHash() (util.Hash, error) {
	return valuehash.NewSHA256(ks.Bytes()), nil
}

func (ks ContractAccountKeys) Bytes() []byte {
	return util.UintToBytes(ks.threshold)
}

func (ks ContractAccountKeys) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, ks.h); err != nil {
		return err
	}

	if len(ks.keys) > 0 {
		return util.ErrInvalid.Errorf("keys of contract account exist")
	}

	if h, err := ks.GenerateHash(); err != nil {
		return err
	} else if !ks.h.Equal(h) {
		return util.ErrInvalid.Errorf("hash not matched")
	}

	return nil
}

func (ks ContractAccountKeys) Threshold() uint {
	return ks.threshold
}

func (ks ContractAccountKeys) Keys() []base.AccountKey {
	return ks.keys
}

func (ks ContractAccountKeys) Key(k mitumbase.Publickey) (base.AccountKey, bool) {
	return base.BaseAccountKey{}, false
}

func (ks ContractAccountKeys) Equal(b base.AccountKeys) bool {
	if ks.threshold != b.Threshold() {
		return false
	}

	if len(ks.keys) != len(b.Keys()) {
		return false
	}

	sort.Slice(ks.keys, func(i, j int) bool {
		return bytes.Compare(ks.keys[i].Key().Bytes(), ks.keys[j].Key().Bytes()) < 0
	})

	bkeys := b.Keys()
	sort.Slice(bkeys, func(i, j int) bool {
		return bytes.Compare(bkeys[i].Key().Bytes(), bkeys[j].Key().Bytes()) < 0
	})

	for i := range ks.keys {
		if !ks.keys[i].Equal(bkeys[i]) {
			return false
		}
	}

	return true
}

func checkThreshold(fs []mitumbase.Sign, keys base.AccountKeys) error {
	var sum uint
	for i := range fs {
		ky, found := keys.Key(fs[i].Signer())
		if !found {
			return errors.Errorf("unknown key found, %s", fs[i].Signer())
		}
		sum += ky.Weight()
	}

	if sum < keys.Threshold() {
		return errors.Errorf("not passed threshold, sum=%d < threshold=%d", sum, keys.Threshold())
	}

	return nil
}
