package currency

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"crypto/ecdsa"
	"crypto/sha256"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

// MEPublickey is the default publickey of mitum, it is based on BTC Privatekey.
type MEPublickey struct {
	k *ecdsa.PublicKey
	s string
	b []byte
	hint.BaseHinter
}

func NewMEPublickey(k *ecdsa.PublicKey) MEPublickey {
	pub := MEPublickey{
		BaseHinter: hint.NewBaseHinter(MEPublickeyHint),
		k:          k,
	}

	return pub.ensure()
}

func ParseMEPublickey(s string) (MEPublickey, error) {
	t := MEPublickeyHint.Type().String()

	switch {
	case !strings.HasSuffix(s, t):
		return MEPublickey{}, util.ErrInvalid.Errorf("unknown publickey string")
	case len(s) <= len(t):
		return MEPublickey{}, util.ErrInvalid.Errorf("invalid publickey string; too short")
	}

	return LoadMEPublickey(s[:len(s)-len(t)])
}

func LoadMEPublickey(s string) (MEPublickey, error) {
	h, err := hex.DecodeString(s)
	if err != nil {
		return MEPublickey{}, util.ErrInvalid.Wrapf(err, "failed to load publickey")
	}
	pk, err := crypto.UnmarshalPubkey(h)
	if err != nil {
		return MEPublickey{}, util.ErrInvalid.Wrapf(err, "failed to unmarshal publickey")
	}

	return NewMEPublickey(pk), nil
}

func (k MEPublickey) String() string {
	return k.s
}

func (k MEPublickey) Bytes() []byte {
	return k.b
}

func (k MEPublickey) IsValid([]byte) error {
	if err := k.BaseHinter.IsValid(MEPublickeyHint.Type().Bytes()); err != nil {
		return util.ErrInvalid.Wrapf(err, "wrong hint in publickey")
	}

	switch {
	case k.k == nil:
		return util.ErrInvalid.Errorf("empty btc publickey in publickey")
	case len(k.s) < 1:
		return util.ErrInvalid.Errorf("empty publickey string")
	case len(k.b) < 1:
		return util.ErrInvalid.Errorf("empty publickey []byte")
	}

	return nil
}

func (k MEPublickey) Equal(b base.PKKey) bool {
	switch {
	case b == nil:
		return false
	default:
		return k.s == b.String()
	}
}

func (k MEPublickey) Verify(input []byte, sig base.Signature) error {
	if len(sig) < 4 {
		return base.ErrSignatureVerification.Call()
	}

	rlength := int(binary.LittleEndian.Uint32(sig[:4]))
	r := big.NewInt(0).SetBytes(sig[4 : 4+rlength])
	s := big.NewInt(0).SetBytes(sig[4+rlength:])

	h := sha256.Sum256(input)
	if !ecdsa.Verify(k.k, h[:], r, s) {
		return base.ErrSignatureVerification.Call()
	}

	return nil
}

func (k MEPublickey) MarshalText() ([]byte, error) {
	return []byte(k.s), nil
}

func (k *MEPublickey) UnmarshalText(b []byte) error {
	u, err := LoadMEPublickey(string(b))
	if err != nil {
		return errors.Wrap(err, "failed to UnmarshalText for publickey")
	}

	*k = u.ensure()

	return nil
}

func (k *MEPublickey) ensure() MEPublickey {
	if k.k == nil {
		return *k
	}

	k.s = fmt.Sprintf("%s%s", hex.EncodeToString(crypto.FromECDSAPub(k.k)), k.Hint().Type().String())
	k.b = []byte(k.s)

	return *k
}
