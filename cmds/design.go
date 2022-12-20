package cmds

import (
	"context"
	"crypto/tls"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/spikeekips/mitum-currency/digest/config"
	"github.com/spikeekips/mitum-currency/digest/util"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/launch"
	mitumutil "github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"

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

type DigestDesignYAMLUnmarshaler struct {
	NetworkYAML  map[string]interface{} `yaml:"network"`
	CacheYAML    string                 `yaml:"cache"`
	DatabaseYAML map[string]interface{} `yaml:"database"`
}

func (d *DigestDesign) DecodeYAML(b []byte, enc *jsonenc.Encoder) error {
	e := mitumutil.StringErrorFunc("failed to unmarshal NodeDesign")

	var u DigestDesignYAMLUnmarshaler

	if err := yaml.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	d.CacheYAML = &u.CacheYAML
	d.NetworkYAML = &LocalNetwork{}
	d.DatabaseYAML = &config.DatabaseYAML{}

	lb, err := mitumutil.MarshalJSON(u.NetworkYAML)
	db, err := mitumutil.MarshalJSON(u.DatabaseYAML)
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

func (no *DigestDesign) Set(ctx context.Context) (context.Context, error) {
	e := mitumutil.StringErrorFunc("failed to Set DigestDesign")

	nctx := context.WithValue(
		context.Background(),
		ContextValueLocalNetwork,
		config.EmptyBaseLocalNetwork(),
	)
	if no.NetworkYAML != nil {
		var conf config.LocalNetwork
		if i, err := no.NetworkYAML.Set(nctx); err != nil {
			return ctx, e(err, "")
		} else if err := mitumutil.LoadFromContext(i, ContextValueLocalNetwork, &conf); err != nil {
			return ctx, e(err, "")
		} else {
			no.network = conf
		}
	}

	var lconf config.LocalNetwork
	if err := mitumutil.LoadFromContext(ctx, ContextValueLocalNetwork, &lconf); err != nil {
		return ctx, e(err, "")
	}

	if no.network.Bind() == nil {
		_ = no.network.SetBind(DefaultDigestAPIBind)
	}

	if no.network.ConnInfo().URL() == nil {
		connInfo, _ := util.NewHTTPConnInfoFromString(DefaultDigestURL, lconf.ConnInfo().Insecure())
		_ = no.network.SetConnInfo(connInfo)
	}

	if certs := no.network.Certs(); len(certs) < 1 {
		priv, err := GenerateED25519Privatekey()
		if err != nil {
			return ctx, e(err, "")
		}

		host := "localhost"
		if no.network.ConnInfo().URL() != nil {
			host = no.network.ConnInfo().URL().Hostname()
		}

		ct, err := GenerateTLSCerts(host, priv)
		if err != nil {
			return ctx, e(err, "")
		}

		if err := no.network.SetCerts(ct); err != nil {
			return ctx, e(err, "")
		}
	}

	if no.CacheYAML == nil {
		no.cache = DefaultDigestAPICache
	} else {
		u, err := util.ParseURL(*no.CacheYAML, true)
		if err != nil {
			return ctx, e(err, "")
		}
		no.cache = u
	}

	var st config.BaseDatabase
	if no.DatabaseYAML == nil {
		if err := st.SetURI(config.DefaultDatabaseURI); err != nil {
			return ctx, e(err, "")
		} else if err := st.SetCache(config.DefaultDatabaseCache); err != nil {
			return ctx, e(err, "")
		} else {
			no.database = st
		}
	} else {
		if err := st.SetURI(no.DatabaseYAML.URI); err != nil {
			return ctx, e(err, "")
		}
		if no.DatabaseYAML.Cache != "" {
			err := st.SetCache(no.DatabaseYAML.Cache)
			if err != nil {
				return ctx, e(err, "")
			}
		}
		no.database = st
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

type NodeDesign struct {
	Address        base.Address
	Privatekey     base.Privatekey
	Storage        launch.NodeStorageDesign
	Network        launch.NodeNetworkDesign
	NetworkID      base.NetworkID
	Digest         DigestDesign
	LocalParams    *isaac.LocalParams
	SyncSources    launch.SyncSourcesDesign
	TimeServerPort int
	TimeServer     string
}

func NodeDesignFromFile(f string, enc *jsonenc.Encoder) (d NodeDesign, _ []byte, _ error) {
	e := mitumutil.StringErrorFunc("failed to load NodeDesign from file")

	b, err := os.ReadFile(filepath.Clean(f))
	if err != nil {
		return d, nil, e(err, "")
	}

	if err := d.DecodeYAML(b, enc); err != nil {
		return d, b, e(err, "")
	}

	return d, b, nil
}

func NodeDesignFromHTTP(u string, tlsInsecure bool, enc *jsonenc.Encoder) (design NodeDesign, _ error) {
	e := mitumutil.StringErrorFunc("failed to load NodeDesign thru http")

	httpclient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: tlsInsecure,
			},
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		return design, e(err, "")
	}

	res, err := httpclient.Do(req)
	if err != nil {
		return design, e(err, "")
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return design, e(err, "")
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return design, e(nil, "design not found")
	}

	if err := design.DecodeYAML(b, enc); err != nil {
		return design, e(err, "")
	}

	return design, nil
}

func NodeDesignFromConsul(addr, key string, enc *jsonenc.Encoder) (design NodeDesign, _ error) {
	e := mitumutil.StringErrorFunc("failed to load NodeDesign thru consul")

	client, err := consulClient(addr)
	if err != nil {
		return design, e(err, "")
	}

	switch v, _, err := client.KV().Get(key, nil); {
	case err != nil:
		return design, e(err, "")
	default:
		if err := design.DecodeYAML(v.Value, enc); err != nil {
			return design, e(err, "")
		}

		return design, nil
	}
}

func (d *NodeDesign) IsValid([]byte) error {
	e := mitumutil.ErrInvalid.Errorf("invalid NodeDesign")

	if len(d.TimeServer) > 0 {
		switch i, err := url.Parse("http://" + d.TimeServer); {
		case err != nil:
			return e.Wrapf(err, "invalid time server, %q", d.TimeServer)
		case len(i.Hostname()) < 1:
			return e.Errorf("invalid time server, %q", d.TimeServer)
		case i.Host != d.TimeServer && len(i.Port()) < 1:
			return e.Errorf("invalid time server, %q", d.TimeServer)
		default:
			s := d.TimeServer
			if len(i.Port()) < 1 {
				s = net.JoinHostPort(d.TimeServer, "123")
			}

			if _, err := net.ResolveUDPAddr("udp", s); err != nil {
				return e.Wrapf(err, "invalid time server, %q", d.TimeServer)
			}

			if len(i.Port()) > 0 {
				p, err := strconv.ParseInt(i.Port(), 10, 64)
				if err != nil {
					return e.Wrapf(err, "invalid time server, %q", d.TimeServer)
				} else if p > 0 && p < math.MaxInt {
					d.TimeServer = i.Hostname()
					d.TimeServerPort = int(p)
				} else {
					return e.Wrapf(err, "invalid time server port, %v", p)
				}
			}
		}
	}

	if err := mitumutil.CheckIsValiders(nil, false, d.Address, d.Privatekey, d.NetworkID); err != nil {
		return e.Wrap(err)
	}

	if err := d.Network.IsValid(nil); err != nil {
		return e.Wrap(err)
	}

	if err := d.Storage.IsValid(nil); err != nil {
		return e.Wrap(err)
	}

	if err := launch.IsValidSyncSourcesDesign(
		d.SyncSources,
		d.Address,
		d.Network.PublishString,
		d.Network.Publish().String(),
	); err != nil {
		return e.Wrap(err)
	}

	switch {
	case d.LocalParams == nil:
		d.LocalParams = isaac.DefaultLocalParams(d.NetworkID)
	default:
		if err := d.LocalParams.IsValid(d.NetworkID); err != nil {
			return e.Wrap(err)
		}
	}

	if err := d.Storage.Patch(d.Address); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (d *NodeDesign) Check(devflags launch.DevFlags) error {
	if !devflags.AllowRiskyThreshold {
		if t := d.LocalParams.Threshold(); t < base.SafeThreshold {
			return mitumutil.ErrInvalid.Errorf("risky threshold under %v; %v", t, base.SafeThreshold)
		}
	}

	return nil
}

type NodeDesignYAMLMarshaler struct {
	Address     base.Address             `yaml:"address"`
	Privatekey  base.Privatekey          `yaml:"privatekey"`
	Storage     launch.NodeStorageDesign `yaml:"storage"`
	NetworkID   string                   `yaml:"network_id"`
	TimeServer  string                   `yaml:"time_server,omitempty"`
	Network     launch.NodeNetworkDesign `yaml:"network"`
	Digest      DigestDesign             `yaml:"digest"`
	LocalParams *isaac.LocalParams       `yaml:"parameters"` //nolint:tagliatelle //...
	SyncSources launch.SyncSourcesDesign `yaml:"sync_sources"`
}

type NodeDesignYAMLUnmarshaler struct {
	SyncSources interface{}                           `yaml:"sync_sources"`
	Storage     launch.NodeStorageDesignYAMLMarshal   `yaml:"storage"`
	Address     string                                `yaml:"address"`
	Privatekey  string                                `yaml:"privatekey"`
	NetworkID   string                                `yaml:"network_id"`
	TimeServer  string                                `yaml:"time_server,omitempty"`
	LocalParams map[string]interface{}                `yaml:"parameters"` //nolint:tagliatelle //...
	Network     launch.NodeNetworkDesignYAMLMarshaler `yaml:"network"`
	Digest      DigestDesign                          `yaml:"digest"`
}

func (d NodeDesign) MarshalYAML() (interface{}, error) {
	return NodeDesignYAMLMarshaler{
		Address:     d.Address,
		Privatekey:  d.Privatekey,
		NetworkID:   string(d.NetworkID),
		Network:     d.Network,
		Storage:     d.Storage,
		LocalParams: d.LocalParams,
		Digest:      d.Digest,
		TimeServer:  d.TimeServer,
	}, nil
}

func (d *NodeDesign) DecodeYAML(b []byte, enc *jsonenc.Encoder) error {
	e := mitumutil.StringErrorFunc("failed to unmarshal NodeDesign")

	var u NodeDesignYAMLUnmarshaler

	if err := yaml.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	switch address, err := base.DecodeAddress(u.Address, enc); {
	case err != nil:
		return e(err, "invalid address")
	default:
		d.Address = address
	}

	switch priv, err := base.DecodePrivatekeyFromString(u.Privatekey, enc); {
	case err != nil:
		return e(err, "invalid privatekey")
	default:
		d.Privatekey = priv
	}

	d.NetworkID = base.NetworkID([]byte(u.NetworkID))

	switch i, err := u.Network.Decode(enc); {
	case err != nil:
		return e(err, "")
	default:
		d.Network = i
	}

	switch i, err := u.Storage.Decode(enc); {
	case err != nil:
		return e(err, "")
	default:
		d.Storage = i
	}

	switch sb, err := yaml.Marshal(u.SyncSources); {
	case err != nil:
		return e(err, "")
	default:
		if err := d.SyncSources.DecodeYAML(sb, enc); err != nil {
			return e(err, "")
		}
	}

	d.LocalParams = isaac.DefaultLocalParams(d.NetworkID)

	switch lb, err := mitumutil.MarshalJSON(u.LocalParams); {
	case err != nil:
		return e(err, "")
	default:
		if err := mitumutil.UnmarshalJSON(lb, d.LocalParams); err != nil {
			return e(err, "")
		}

		d.LocalParams.BaseHinter = hint.NewBaseHinter(isaac.LocalParamsHint)
	}

	d.TimeServer = strings.TrimSpace(u.TimeServer)

	dd, err := mitumutil.MarshalJSON(u.Digest)
	if err != nil {
		return e(err, "")
	}

	if err := yaml.Unmarshal(dd, &u); err != nil {
		return e(err, "")
	}

	if (u.Digest != DigestDesign{}) {
		d.Digest = u.Digest
		d.Digest.CacheYAML = u.Digest.CacheYAML
		d.Digest.NetworkYAML = &LocalNetwork{}

		switch lb, err := mitumutil.MarshalJSON(u.Digest.NetworkYAML); {
		case err != nil:
			return e(err, "")
		default:
			if err := mitumutil.UnmarshalJSON(lb, d.Digest.NetworkYAML); err != nil {
				return e(err, "")
			}
		}
	}

	return nil
}

func (d NodeDesign) MarshalZerologObject(e *zerolog.Event) {
	var priv base.Publickey
	if d.Privatekey != nil {
		priv = d.Privatekey.Publickey()
	}

	e.
		Interface("address", d.Address).
		Interface("privatekey*", priv).
		Interface("storage", d.Storage).
		Interface("network_id", d.NetworkID).
		Interface("network", d.Network).
		Interface("parameters", d.LocalParams).
		Interface("sync_sources", d.SyncSources)
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
