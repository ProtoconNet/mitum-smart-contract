package types

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

var (
	MEPrivatekeyHint = hint.MustNewHint("fpr-v0.0.1")
	MEPublickeyHint  = hint.MustNewHint("fpu-v0.0.1")
)

// MEPrivatekey is the default privatekey of mitum, it is based on BTC Privatekey.
type MEPrivatekey struct {
	priv *ecdsa.PrivateKey
	s    string
	pub  MEPublickey
	b    []byte
	hint.BaseHinter
}

func NewMEPrivatekey() MEPrivatekey {
	priv, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)

	return newMEPrivatekeyFromPrivateKey(priv)
}

func NewMEPrivatekeyFromSeed(s string) (MEPrivatekey, error) {
	if l := len([]byte(s)); l < base.PrivatekeyMinSeedSize {
		return MEPrivatekey{}, util.ErrInvalid.Errorf(
			"wrong seed for privatekey; too short, %d < %d", l, base.PrivatekeyMinSeedSize)
	}

	priv, err := ecdsa.GenerateKey(
		btcec.S256(),
		bytes.NewReader([]byte(valuehash.NewSHA256([]byte(s)).String())),
	)
	if err != nil {
		return MEPrivatekey{}, errors.WithStack(err)
	}

	return newMEPrivatekeyFromPrivateKey(priv), nil
}

func ParseMEPrivatekey(s string) (MEPrivatekey, error) {
	t := MEPrivatekeyHint.Type().String()

	switch {
	case !strings.HasSuffix(s, t):
		return MEPrivatekey{}, util.ErrInvalid.Errorf("Unknown private key string")
	case len(s) <= len(t):
		return MEPrivatekey{}, util.ErrInvalid.Errorf("Invalid private key string; too short")
	}

	return LoadMEPrivatekey(s[:len(s)-len(t)])
}

func LoadMEPrivatekey(s string) (MEPrivatekey, error) {
	h, err := hex.DecodeString(s)
	if err != nil {
		return MEPrivatekey{}, err
	}

	priv, err := crypto.ToECDSA(h)
	if err != nil {
		return MEPrivatekey{}, err
	}

	return newMEPrivatekeyFromPrivateKey(priv), nil
}

func newMEPrivatekeyFromPrivateKey(priv *ecdsa.PrivateKey) MEPrivatekey {
	k := MEPrivatekey{
		BaseHinter: hint.NewBaseHinter(MEPrivatekeyHint),
		priv:       priv,
	}

	return k.ensure()
}

func (k MEPrivatekey) String() string {
	return k.s
}

func (k MEPrivatekey) Bytes() []byte {
	return k.b
}

func (k MEPrivatekey) IsValid([]byte) error {
	if err := k.BaseHinter.IsValid(MEPrivatekeyHint.Type().Bytes()); err != nil {
		return util.ErrInvalid.WithMessage(err, "wrong hint in private key")
	}

	switch {
	case k.priv == nil:
		return util.ErrInvalid.Errorf("Empty btc private key")
	case len(k.s) < 1:
		return util.ErrInvalid.Errorf("Empty private key string")
	case len(k.b) < 1:
		return util.ErrInvalid.Errorf("Empty private key []byte")
	}

	return nil
}

func (k MEPrivatekey) Publickey() base.Publickey {
	return k.pub
}

func (k MEPrivatekey) Equal(b base.PKKey) bool {
	switch {
	case b == nil:
		return false
	default:
		return k.s == b.String()
	}
}

func (k MEPrivatekey) Sign(b []byte) (base.Signature, error) {
	h := sha256.Sum256(b)
	r, s, err := ecdsa.Sign(rand.Reader, k.priv, h[:])
	if err != nil {
		return nil, err
	}

	bs := make([]byte, 4+len(r.Bytes())+len(s.Bytes()))
	binary.LittleEndian.PutUint32(bs, uint32(len(r.Bytes())))

	copy(bs[4:], r.Bytes())
	copy(bs[4+len(r.Bytes()):], s.Bytes())
	return base.Signature(bs), nil
}

func (k MEPrivatekey) MarshalText() ([]byte, error) {
	return []byte(k.s), nil
}

func (k *MEPrivatekey) UnmarshalText(b []byte) error {
	u, err := LoadMEPrivatekey(string(b))
	if err != nil {
		return err
	}

	*k = u.ensure()

	return nil
}

func (k *MEPrivatekey) ensure() MEPrivatekey {
	if k.priv == nil {
		return *k
	}

	k.pub = NewMEPublickey(&k.priv.PublicKey)
	k.s = fmt.Sprintf("%s%s", hex.EncodeToString(crypto.FromECDSA(k.priv)), k.Hint().Type().String())
	k.b = []byte(k.s)

	return *k
}
