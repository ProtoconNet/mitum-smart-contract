package cmds

import (
	"bytes"
	"context"
	"net/url"
	"sort"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"

	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/ProtoconNet/mitum-currency/v3/digest/config"
	"github.com/ProtoconNet/mitum-currency/v3/digest/util"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/network/quicstream"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

var (
	DefaultDigestAPICache *url.URL
	DefaultDigestAPIBind  string
	//DefaultDigestAPIURL   string
)

func init() {
	DefaultDigestAPICache, _ = util.ParseURL("memory://", false)
	DefaultDigestAPIBind = "https://0.0.0.0:54320"
	//DefaultDigestAPIURL = "https://127.0.0.1:54320"
}

type KeyDesign struct {
	PublicKeyString string `yaml:"publickey"`
	Weight          uint
	Key             types.BaseAccountKey `yaml:"-"`
}

func (kd *KeyDesign) IsValid([]byte) error {
	je, _ := encs.Find(jsonenc.JSONEncoderHint)

	if pub, err := base.DecodePublickeyFromString(kd.PublicKeyString, je); err != nil {
		return mitumutil.ErrInvalid.Wrap(err)
	} else if k, err := types.NewBaseAccountKey(pub, kd.Weight); err != nil {
		return mitumutil.ErrInvalid.Wrap(err)
	} else {
		kd.Key = k
	}

	return nil
}

type AccountKeysDesign struct {
	Threshold  uint
	KeysDesign []*KeyDesign      `yaml:"keys"`
	Keys       types.AccountKeys `yaml:"-"`
	Address    types.Address     `yaml:"-"`
}

func (akd *AccountKeysDesign) IsValid([]byte) error {
	ks := make([]types.AccountKey, len(akd.KeysDesign))
	for i := range akd.KeysDesign {
		kd := akd.KeysDesign[i]

		if err := kd.IsValid(nil); err != nil {
			return err
		}

		ks[i] = kd.Key
	}

	keys, err := types.NewBaseAccountKeys(ks, akd.Threshold)
	if err != nil {
		return mitumutil.ErrInvalid.Wrap(err)
	}
	akd.Keys = keys

	a, err := types.NewAddressFromKeys(akd.Keys)
	if err != nil {
		return mitumutil.ErrInvalid.Wrap(err)
	}
	akd.Address = a

	return nil
}

type GenesisCurrencyDesign struct {
	AccountKeys *AccountKeysDesign `yaml:"account-keys"`
	Currencies  []*CurrencyDesign  `yaml:"currencies"`
}

func (de *GenesisCurrencyDesign) IsValid([]byte) error {
	if de.AccountKeys == nil {
		return errors.Errorf("empty account-keys")
	}

	if err := de.AccountKeys.IsValid(nil); err != nil {
		return err
	}

	for i := range de.Currencies {
		if err := de.Currencies[i].IsValid(nil); err != nil {
			return err
		}
	}

	return nil
}

type CurrencyDesign struct {
	CurrencyString             *string      `yaml:"currency"`
	BalanceString              *string      `yaml:"balance"`
	NewAccountMinBalanceString *string      `yaml:"new-account-min-balance"`
	Feeer                      *FeeerDesign `yaml:"feeer"`
	Balance                    types.Amount `yaml:"-"`
	NewAccountMinBalance       common.Big   `yaml:"-"`
}

func (de *CurrencyDesign) IsValid([]byte) error {
	var cid types.CurrencyID
	if de.CurrencyString == nil {
		return errors.Errorf("empty currency")
	}
	cid = types.CurrencyID(*de.CurrencyString)
	if err := cid.IsValid(nil); err != nil {
		return err
	}

	if de.BalanceString != nil {
		b, err := common.NewBigFromString(*de.BalanceString)
		if err != nil {
			return mitumutil.ErrInvalid.Wrap(err)
		}
		de.Balance = types.NewAmount(b, cid)
		if err := de.Balance.IsValid(nil); err != nil {
			return err
		}
	}

	if de.NewAccountMinBalanceString == nil {
		de.NewAccountMinBalance = common.ZeroBig
	} else {
		b, err := common.NewBigFromString(*de.NewAccountMinBalanceString)
		if err != nil {
			return mitumutil.ErrInvalid.Wrap(err)
		}
		de.NewAccountMinBalance = b
	}

	if de.Feeer == nil {
		de.Feeer = &FeeerDesign{}
	} else if err := de.Feeer.IsValid(nil); err != nil {
		return err
	}

	return nil
}

// FeeerDesign is used for genesis currencies and naturally it's receiver is genesis account
type FeeerDesign struct {
	Type   string
	Extras map[string]interface{} `yaml:",inline"`
}

func (no *FeeerDesign) IsValid([]byte) error {
	switch t := no.Type; t {
	case types.FeeerNil, "":
	case types.FeeerFixed:
		if err := no.checkFixed(no.Extras); err != nil {
			return err
		}
	case types.FeeerRatio:
		if err := no.checkRatio(no.Extras); err != nil {
			return err
		}
	default:
		return errors.Errorf("unknown type of feeer, %v", t)
	}

	return nil
}

func (no FeeerDesign) checkFixed(c map[string]interface{}) error {
	a, found := c["amount"]
	if !found {
		return errors.Errorf("fixed needs `amount`")
	}
	n, err := common.NewBigFromInterface(a)
	if err != nil {
		return errors.Wrapf(err, "invalid amount value, %v of fixed", a)
	}
	no.Extras["fixed_amount"] = n

	return nil
}

func (no FeeerDesign) checkRatio(c map[string]interface{}) error {
	if a, found := c["ratio"]; !found {
		return errors.Errorf("ratio needs `ratio`")
	} else if f, ok := a.(float64); !ok {
		return errors.Errorf("invalid ratio value type, %T of ratio; should be float64", a)
	} else {
		no.Extras["ratio_ratio"] = f
	}

	if a, found := c["min"]; !found {
		return errors.Errorf("ratio needs `min`")
	} else if n, err := common.NewBigFromInterface(a); err != nil {
		return errors.Wrapf(err, "invalid min value, %v of ratio", a)
	} else {
		no.Extras["ratio_min"] = n
	}

	if a, found := c["max"]; found {
		n, err := common.NewBigFromInterface(a)
		if err != nil {
			return errors.Wrapf(err, "invalid max value, %v of ratio", a)
		}
		no.Extras["ratio_max"] = n
	}

	return nil
}

type DigestDesign struct {
	NetworkYAML  *LocalNetwork         `yaml:"network,omitempty"`
	CacheYAML    *string               `yaml:"cache,omitempty"`
	DatabaseYAML *config.DatabaseYAML  `yaml:"database"`
	ConnInfo     []quicstream.ConnInfo `yaml:"conn_info,omitempty"`
	network      config.LocalNetwork
	database     config.BaseDatabase
	cache        *url.URL
}

func (d *DigestDesign) Set(ctx context.Context) (context.Context, error) {
	e := mitumutil.StringError("set DigestDesign")

	nctx := context.WithValue(
		context.Background(),
		ContextValueLocalNetwork,
		config.EmptyBaseLocalNetwork(),
	)
	p := &LocalNetwork{}
	if *d.NetworkYAML != *p {
		var conf config.LocalNetwork
		if i, err := d.NetworkYAML.Set(nctx); err != nil {
			return ctx, e.Wrap(err)
		} else if err := mitumutil.LoadFromContext(i, ContextValueLocalNetwork, &conf); err != nil {
			return ctx, e.Wrap(err)
		} else {
			d.network = conf
		}
	}

	var ndesign launch.NodeDesign
	if err := mitumutil.LoadFromContext(ctx, launch.DesignContextKey, &ndesign); err != nil {
		return ctx, err
	}

	if d.network.Bind() == nil {
		_ = d.network.SetBind(DefaultDigestAPIBind)
	}

	if d.network.ConnInfo().URL() == nil {
		connInfo, _ := util.NewHTTPConnInfoFromString(DefaultDigestURL, ndesign.Network.TLSInsecure)
		_ = d.network.SetConnInfo(connInfo)
	}

	if certs := d.network.Certs(); len(certs) < 1 {
		priv, err := GenerateED25519PrivateKey()
		if err != nil {
			return ctx, e.Wrap(err)
		}

		host := "localhost"
		if d.network.ConnInfo().URL() != nil {
			host = d.network.ConnInfo().URL().Hostname()
		}

		ct, err := GenerateTLSCerts(host, priv)
		if err != nil {
			return ctx, e.Wrap(err)
		}

		if err := d.network.SetCerts(ct); err != nil {
			return ctx, e.Wrap(err)
		}
	}

	if d.CacheYAML == nil {
		d.cache = DefaultDigestAPICache
	} else {
		u, err := util.ParseURL(*d.CacheYAML, true)
		if err != nil {
			return ctx, e.Wrap(err)
		}
		d.cache = u
	}

	var st config.BaseDatabase
	if d.DatabaseYAML == nil {
		if err := st.SetURI(config.DefaultDatabaseURI); err != nil {
			return ctx, e.Wrap(err)
		} else if err := st.SetCache(config.DefaultDatabaseCache); err != nil {
			return ctx, e.Wrap(err)
		} else {
			d.database = st
		}
	} else {
		if err := st.SetURI(d.DatabaseYAML.URI); err != nil {
			return ctx, e.Wrap(err)
		}
		if d.DatabaseYAML.Cache != "" {
			err := st.SetCache(d.DatabaseYAML.Cache)
			if err != nil {
				return ctx, e.Wrap(err)
			}
		}
		d.database = st
	}

	return ctx, nil
}

func (d *DigestDesign) Network() config.LocalNetwork {
	return d.network
}

func (d *DigestDesign) Cache() *url.URL {
	return d.cache
}

func (d *DigestDesign) Database() config.BaseDatabase {
	return d.database
}

func (d DigestDesign) MarshalZerologObject(e *zerolog.Event) {
	e.
		Interface("network", d.network).
		Interface("database", d.database).
		Interface("cache", d.cache)
}

func (d DigestDesign) Equal(b DigestDesign) bool {
	if d.NetworkYAML != b.NetworkYAML {
		return false
	}

	if d.CacheYAML != b.CacheYAML {
		return false
	}

	if d.DatabaseYAML != b.DatabaseYAML {
		return false
	}

	if len(d.ConnInfo) != len(b.ConnInfo) {
		return false
	}

	sort.Slice(d.ConnInfo, func(i, j int) bool {
		return bytes.Compare([]byte(d.ConnInfo[i].String()), []byte(d.ConnInfo[j].String())) < 0
	})

	bConn := b.ConnInfo
	sort.Slice(bConn, func(i, j int) bool {
		return bytes.Compare([]byte(bConn[i].String()), []byte(bConn[j].String())) < 0
	})

	for i := range d.ConnInfo {
		if !(d.ConnInfo[i].String() == bConn[i].String()) {
			return false
		}
	}

	return true
}

func loadPrivatekeyFromVault(path string, enc *jsonenc.Encoder) (base.Privatekey, error) {
	e := mitumutil.StringError("load private key from vault")

	clientConfig := vault.DefaultConfig()

	client, err := vault.NewClient(clientConfig)
	if err != nil {
		return nil, e.WithMessage(err, "create vault client")
	}

	secret, err := client.KVv2("secret").Get(context.Background(), path)
	if err != nil {
		return nil, e.WithMessage(err, "to read secret")
	}

	i := secret.Data["string"]

	privs, ok := i.(string)
	if !ok {
		return nil, e.WithMessage(nil, "read secret; expected string but %T", i)
	}

	switch priv, err := base.DecodePrivatekeyFromString(privs, enc); {
	case err != nil:
		return nil, e.WithMessage(err, "invalid privatekey")
	default:
		return priv, nil
	}
}
