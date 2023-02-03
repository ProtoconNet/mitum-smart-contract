package cmds

import (
	"context"
	"net/url"
	"os"
	"path/filepath"

	consulapi "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/spikeekips/mitum-currency/digest/config"
	"github.com/spikeekips/mitum-currency/digest/util"
	"github.com/spikeekips/mitum/base"
	mitumutil "github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"

	"github.com/spikeekips/mitum-currency/currency"
	"gopkg.in/yaml.v3"
)

var (
	DefaultDigestAPICache *url.URL
	DefaultDigestAPIBind  string
	DefaultDigestAPIURL   string
)

func init() {
	DefaultDigestAPICache, _ = util.ParseURL("memory://", false)
	DefaultDigestAPIBind = "https://0.0.0.0:54320"
	DefaultDigestAPIURL = "https://127.0.0.1:54320"
}

type KeyDesign struct {
	PublickeyString string `yaml:"publickey"`
	Weight          uint
	Key             currency.BaseAccountKey `yaml:"-"`
}

func (kd *KeyDesign) IsValid([]byte) error {
	je := encs.Find(jsonenc.JSONEncoderHint)

	if pub, err := base.DecodePublickeyFromString(kd.PublickeyString, je); err != nil {
		return mitumutil.ErrInvalid.Wrap(err)
	} else if k, err := currency.NewBaseAccountKey(pub, kd.Weight); err != nil {
		return mitumutil.ErrInvalid.Wrap(err)
	} else {
		kd.Key = k
	}

	return nil
}

type AccountKeysDesign struct {
	Threshold  uint
	KeysDesign []*KeyDesign             `yaml:"keys"`
	Keys       currency.BaseAccountKeys `yaml:"-"`
	Address    currency.Address         `yaml:"-"`
}

func (akd *AccountKeysDesign) IsValid([]byte) error {
	ks := make([]currency.AccountKey, len(akd.KeysDesign))
	for i := range akd.KeysDesign {
		kd := akd.KeysDesign[i]

		if err := kd.IsValid(nil); err != nil {
			return err
		}

		ks[i] = kd.Key
	}

	keys, err := currency.NewBaseAccountKeys(ks, akd.Threshold)
	if err != nil {
		return mitumutil.ErrInvalid.Wrap(err)
	}
	akd.Keys = keys

	a, err := currency.NewAddressFromKeys(akd.Keys)
	if err != nil {
		return mitumutil.ErrInvalid.Wrap(err)
	}
	akd.Address = a

	return nil
}

type GenesisCurrenciesDesign struct {
	AccountKeys *AccountKeysDesign `yaml:"account-keys"`
	Currencies  []*CurrencyDesign  `yaml:"currencies"`
}

func (de *GenesisCurrenciesDesign) IsValid([]byte) error {
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
	CurrencyString             *string         `yaml:"currency"`
	BalanceString              *string         `yaml:"balance"`
	NewAccountMinBalanceString *string         `yaml:"new-account-min-balance"`
	Feeer                      *FeeerDesign    `yaml:"feeer"`
	Balance                    currency.Amount `yaml:"-"`
	NewAccountMinBalance       currency.Big    `yaml:"-"`
}

func (de *CurrencyDesign) IsValid([]byte) error {
	var cid currency.CurrencyID
	if de.CurrencyString == nil {
		return errors.Errorf("empty currency")
	}
	cid = currency.CurrencyID(*de.CurrencyString)
	if err := cid.IsValid(nil); err != nil {
		return err
	}

	if de.BalanceString != nil {
		b, err := currency.NewBigFromString(*de.BalanceString)
		if err != nil {
			return mitumutil.ErrInvalid.Wrap(err)
		}
		de.Balance = currency.NewAmount(b, cid)
		if err := de.Balance.IsValid(nil); err != nil {
			return err
		}
	}

	if de.NewAccountMinBalanceString == nil {
		de.NewAccountMinBalance = currency.ZeroBig
	} else {
		b, err := currency.NewBigFromString(*de.NewAccountMinBalanceString)
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
	case currency.FeeerNil, "":
	case currency.FeeerFixed:
		if err := no.checkFixed(no.Extras); err != nil {
			return err
		}
	case currency.FeeerRatio:
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
	n, err := currency.NewBigFromInterface(a)
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
	} else if n, err := currency.NewBigFromInterface(a); err != nil {
		return errors.Wrapf(err, "invalid min value, %v of ratio", a)
	} else {
		no.Extras["ratio_min"] = n
	}

	if a, found := c["max"]; found {
		n, err := currency.NewBigFromInterface(a)
		if err != nil {
			return errors.Wrapf(err, "invalid max value, %v of ratio", a)
		}
		no.Extras["ratio_max"] = n
	}

	return nil
}

type DigestDesign struct {
	NetworkYAML  *LocalNetwork        `yaml:"network,omitempty"`
	CacheYAML    *string              `yaml:"cache,omitempty"`
	DatabaseYAML *config.DatabaseYAML `yaml:"database"`
	network      config.LocalNetwork
	database     config.BaseDatabase
	cache        *url.URL
}

type DigestYAMLUnmarshaler struct {
	Design DesignYAMLUnmarshaler `yaml:"digest"`
}

type DesignYAMLUnmarshaler struct {
	NetworkYAML  map[string]interface{} `yaml:"network"`
	CacheYAML    *string                `yaml:"cache"`
	DatabaseYAML map[string]interface{} `yaml:"database"`
}

func (d *DigestDesign) DecodeYAML(b []byte, enc *jsonenc.Encoder) error {
	e := mitumutil.StringErrorFunc("failed to unmarshal DigestDesign")

	var u DigestYAMLUnmarshaler

	if err := yaml.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	d.CacheYAML = u.Design.CacheYAML
	d.NetworkYAML = &LocalNetwork{}
	d.DatabaseYAML = &config.DatabaseYAML{}

	lb, err := mitumutil.MarshalJSON(u.Design.NetworkYAML)
	db, err := mitumutil.MarshalJSON(u.Design.DatabaseYAML)
	switch {
	case err != nil:
		return e(err, "")
	default:
		if err := mitumutil.UnmarshalJSON(lb, d.NetworkYAML); err != nil {
			return e(err, "")
		} else if err := mitumutil.UnmarshalJSON(db, d.DatabaseYAML); err != nil {
			return e(err, "")
		}
	}

	return nil
}

func DigestDesignFromFile(f string, enc *jsonenc.Encoder) (d DigestDesign, _ []byte, _ error) {
	e := mitumutil.StringErrorFunc("failed to load DigestDesign from file")

	b, err := os.ReadFile(filepath.Clean(f))
	if err != nil {
		return d, nil, e(err, "")
	}

	if err := d.DecodeYAML(b, enc); err != nil {
		return d, b, e(err, "")
	}

	return d, b, nil
}

func (d *DigestDesign) Set(ctx context.Context) (context.Context, error) {
	e := mitumutil.StringErrorFunc("failed to Set DigestDesign")

	nctx := context.WithValue(
		context.Background(),
		ContextValueLocalNetwork,
		config.EmptyBaseLocalNetwork(),
	)
	if d.NetworkYAML != nil {
		var conf config.LocalNetwork
		if i, err := d.NetworkYAML.Set(nctx); err != nil {
			return ctx, e(err, "")
		} else if err := mitumutil.LoadFromContext(i, ContextValueLocalNetwork, &conf); err != nil {
			return ctx, e(err, "")
		} else {
			d.network = conf
		}
	}

	var lconf config.LocalNetwork
	if err := mitumutil.LoadFromContext(ctx, ContextValueLocalNetwork, &lconf); err != nil {
		return ctx, e(err, "")
	}

	if d.network.Bind() == nil {
		_ = d.network.SetBind(DefaultDigestAPIBind)
	}

	if d.network.ConnInfo().URL() == nil {
		connInfo, _ := util.NewHTTPConnInfoFromString(DefaultDigestURL, lconf.ConnInfo().Insecure())
		_ = d.network.SetConnInfo(connInfo)
	}

	if certs := d.network.Certs(); len(certs) < 1 {
		priv, err := GenerateED25519Privatekey()
		if err != nil {
			return ctx, e(err, "")
		}

		host := "localhost"
		if d.network.ConnInfo().URL() != nil {
			host = d.network.ConnInfo().URL().Hostname()
		}

		ct, err := GenerateTLSCerts(host, priv)
		if err != nil {
			return ctx, e(err, "")
		}

		if err := d.network.SetCerts(ct); err != nil {
			return ctx, e(err, "")
		}
	}

	if d.CacheYAML == nil {
		d.cache = DefaultDigestAPICache
	} else {
		u, err := util.ParseURL(*d.CacheYAML, true)
		if err != nil {
			return ctx, e(err, "")
		}
		d.cache = u
	}

	var st config.BaseDatabase
	if d.DatabaseYAML == nil {
		if err := st.SetURI(config.DefaultDatabaseURI); err != nil {
			return ctx, e(err, "")
		} else if err := st.SetCache(config.DefaultDatabaseCache); err != nil {
			return ctx, e(err, "")
		} else {
			d.database = st
		}
	} else {
		if err := st.SetURI(d.DatabaseYAML.URI); err != nil {
			return ctx, e(err, "")
		}
		if d.DatabaseYAML.Cache != "" {
			err := st.SetCache(d.DatabaseYAML.Cache)
			if err != nil {
				return ctx, e(err, "")
			}
		}
		d.database = st
	}

	return ctx, nil
}

func (no *DigestDesign) Network() config.LocalNetwork {
	return no.network
}

func (no *DigestDesign) Cache() *url.URL {
	return no.cache
}

func (no *DigestDesign) Database() config.BaseDatabase {
	return no.database
}

func (d DigestDesign) MarshalZerologObject(e *zerolog.Event) {
	e.
		Interface("network", d.network).
		Interface("database", d.database).
		Interface("cache", d.cache)
}

func loadPrivatekeyFromVault(path string, enc *jsonenc.Encoder) (base.Privatekey, error) {
	e := mitumutil.StringErrorFunc("failed to load privatekey from vault")

	config := vault.DefaultConfig()

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, e(err, "failed to create vault client")
	}

	secret, err := client.KVv2("secret").Get(context.Background(), path)
	if err != nil {
		return nil, e(err, "failed to read secret")
	}

	i := secret.Data["string"]

	privs, ok := i.(string)
	if !ok {
		return nil, e(nil, "failed to read secret; expected string but %T", i)
	}

	switch priv, err := base.DecodePrivatekeyFromString(privs, enc); {
	case err != nil:
		return nil, e(err, "invalid privatekey")
	default:
		return priv, nil
	}
}

func consulClient(addr string) (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	if len(addr) > 0 {
		config.Address = addr
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new consul api Client")
	}

	return client, nil
}
