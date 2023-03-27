package config

import (
	"crypto/tls"
	"net/url"

	"github.com/ProtoconNet/mitum-currency/v2/digest/cache"
	"github.com/ProtoconNet/mitum-currency/v2/digest/util"
)

var (
	DefaultLocalNetworkURL       = &url.URL{Scheme: "https", Host: "127.0.0.1:54321"}
	DefaultLocalNetworkBind      = &url.URL{Scheme: "https", Host: "0.0.0.0:54321"}
	DefaultLocalNetworkCache     = "gcache:?type=lru&size=100&expire=3s"
	DefaultLocalNetworkSealCache = "gcache:?type=lru&size=10000&expire=3m"
)

type BaseNodeNetwork struct {
	connInfo util.HTTPConnInfo
}

func EmptyBaseNodeNetwork() *BaseNodeNetwork {
	return &BaseNodeNetwork{}
}

func (no BaseNodeNetwork) ConnInfo() util.HTTPConnInfo {
	return no.connInfo
}

func (no *BaseNodeNetwork) SetConnInfo(c util.HTTPConnInfo) error {
	if err := c.IsValid(nil); err != nil {
		return err
	}

	no.connInfo = c

	return nil
}

type LocalNetwork interface {
	ConnInfo() util.HTTPConnInfo
	SetConnInfo(util.HTTPConnInfo) error
	Bind() *url.URL
	SetBind(string) error
	Certs() []tls.Certificate
	SetCerts([]tls.Certificate) error
	Cache() *url.URL
	SetCache(string) error
	SealCache() *url.URL
	SetSealCache(string) error
}

type BaseLocalNetwork struct {
	*BaseNodeNetwork
	bind      *url.URL
	certs     []tls.Certificate
	cache     *url.URL
	sealCache *url.URL
}

func EmptyBaseLocalNetwork() *BaseLocalNetwork {
	return &BaseLocalNetwork{BaseNodeNetwork: &BaseNodeNetwork{}}
}

func (no BaseLocalNetwork) Bind() *url.URL {
	return no.bind
}

func (no *BaseLocalNetwork) SetBind(s string) error {
	u, err := util.ParseURL(s, true)
	if err != nil {
		return err
	}
	no.bind = u

	return nil
}

func (no BaseLocalNetwork) Certs() []tls.Certificate {
	return no.certs
}

func (no *BaseLocalNetwork) SetCerts(certs []tls.Certificate) error {
	no.certs = certs

	return nil
}

func (no BaseLocalNetwork) Cache() *url.URL {
	return no.cache
}

func (no *BaseLocalNetwork) SetCache(s string) error {
	if u, err := util.ParseURL(s, true); err != nil {
		return err
	} else if _, err := cache.NewCacheFromURI(u.String()); err != nil {
		return err
	} else {
		no.cache = u

		return nil
	}
}

func (no BaseLocalNetwork) SealCache() *url.URL {
	return no.sealCache
}

func (no *BaseLocalNetwork) SetSealCache(s string) error {
	if u, err := util.ParseURL(s, true); err != nil {
		return err
	} else if _, err := cache.NewCacheFromURI(u.String()); err != nil {
		return err
	} else {
		no.sealCache = u

		return nil
	}
}
