package config

import (
	"fmt"
	"net/url"

	"github.com/spikeekips/mitum-currency/digest/cache"
	"github.com/spikeekips/mitum-currency/digest/util"
)

var (
	DefaultDatabaseURI   = "mongodb://127.0.0.1:27017/mitum"
	DefaultDatabaseCache = fmt.Sprintf(
		"gcache:?type=%s&size=%d&expire=%s",
		cache.DefaultGCacheType,
		cache.DefaultGCacheSize,
		cache.DefaultCacheExpire.String(),
	)
	DefaultDatabaseCacheURL *url.URL
)

func init() {
	if i, err := util.ParseURL(DefaultDatabaseCache, false); err != nil {
		panic(err)
	} else {
		DefaultDatabaseCacheURL = i
	}
}

type Database interface {
	URI() *url.URL
	SetURI(string) error
	Cache() *url.URL
	SetCache(string) error
}

type BaseDatabase struct {
	uri   *url.URL
	cache *url.URL
}

func (no BaseDatabase) URI() *url.URL {
	return no.uri
}

func (no *BaseDatabase) SetURI(s string) error {
	u, err := util.ParseURL(s, true)
	if err != nil {
		return err
	}
	no.uri = u

	return nil
}

func (no BaseDatabase) Cache() *url.URL {
	return no.cache
}

func (no *BaseDatabase) SetCache(s string) error {
	if u, err := util.ParseURL(s, true); err != nil {
		return err
	} else if _, err := cache.NewCacheFromURI(u.String()); err != nil {
		return err
	} else {
		no.cache = u

		return nil
	}
}

type DatabaseYAML struct {
	URI   string `yaml:",omitempty"`
	Cache string `yaml:",omitempty"`
}
