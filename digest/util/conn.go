package util

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var (
	HTTPConnInfoHint = hint.MustNewHint("http-conninfo-v0.0.1")
)

type HTTPConnInfo struct {
	hint.BaseHinter
	u           *url.URL
	tlsinsecure bool
}

func NewHTTPConnInfo(u *url.URL, insecure bool) HTTPConnInfo {
	return HTTPConnInfo{
		BaseHinter: hint.NewBaseHinter(HTTPConnInfoHint),
		u:          NormalizeURL(u), tlsinsecure: insecure,
	}
}

func NewHTTPConnInfoFromString(s string, insecure bool) (HTTPConnInfo, error) {
	u, err := NormalizeURLString(s)
	if err != nil {
		return HTTPConnInfo{}, errors.Wrapf(err, "wrong node url, %v", s)
	}
	return NewHTTPConnInfo(u, insecure), nil
}

func (conn HTTPConnInfo) IsValid([]byte) error {
	return IsValidURL(conn.u)
}

func (conn HTTPConnInfo) URL() *url.URL {
	return conn.u
}

func (conn HTTPConnInfo) Insecure() bool {
	return conn.tlsinsecure
}

func (conn HTTPConnInfo) SetInsecure(i bool) HTTPConnInfo {
	conn.tlsinsecure = i

	return conn
}

func (conn HTTPConnInfo) Bytes() []byte {
	var v int8
	if conn.tlsinsecure {
		v = 1
	}
	return util.ConcatBytesSlice(
		[]byte(conn.u.String()),
		[]byte{byte(v)},
	)
}

func (conn HTTPConnInfo) String() string {
	s := conn.u.String()
	if conn.tlsinsecure {
		s += "#insecure"
	}

	return s
}

func CheckBindIsOpen(network, bind string, timeout time.Duration) error {
	errchan := make(chan error)
	switch network {
	case "tcp":
		go func() {
			if server, err := net.Listen(network, bind); err != nil {
				errchan <- err
			} else if server != nil {
				_ = server.Close()
			}
		}()
	case "udp":
		go func() {
			if server, err := net.ListenPacket(network, bind); err != nil {
				errchan <- err
			} else if server != nil {
				_ = server.Close()
			}
		}()
	}

	select {
	case err := <-errchan:
		return errors.Wrap(err, "open bind")
	case <-time.After(timeout):
		return nil
	}
}

func ParseURL(s string, allowEmpty bool) (*url.URL, error) { // nolint:unparam
	s = strings.TrimSpace(s)
	if len(s) < 1 {
		if !allowEmpty {
			return nil, errors.Errorf("empty url string")
		}

		return nil, nil
	}

	return url.Parse(s)
}

func NormalizeURLString(s string) (*url.URL, error) {
	u, err := ParseURL(s, false)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid url, %v", s)
	}

	return NormalizeURL(u), nil
}

func NormalizeURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}

	uu := &url.URL{
		Scheme:      u.Scheme,
		Opaque:      u.Opaque,
		User:        u.User,
		Host:        u.Host,
		Path:        u.Path,
		RawPath:     u.RawPath,
		ForceQuery:  u.ForceQuery,
		RawQuery:    u.RawQuery,
		Fragment:    u.Fragment,
		RawFragment: u.RawFragment,
	}

	hostname := uu.Hostname()
	if strings.EqualFold(uu.Hostname(), "localhost") {
		hostname = "127.0.0.1"
	}

	port := uu.Port()
	if port == "" {
		switch uu.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			port = "0"
		}
	}

	uu.Host = hostname + ":" + port

	if uu.Path == "/" {
		uu.Path = ""
	}

	return uu
}

func IsValidURL(u *url.URL) error {
	if u == nil {
		return util.ErrInvalid.Errorf("empty url")
	}
	if u.Scheme == "" {
		return util.ErrInvalid.Errorf("empty scheme, %v", u.String())
	}

	switch {
	case u.Host == "":
		return util.ErrInvalid.Errorf("empty host, %v", u.String())
	case strings.HasPrefix(u.Host, ":") && u.Host == fmt.Sprintf(":%s", u.Port()):
		return util.ErrInvalid.Errorf("empty host, %v", u.String())
	}

	return nil
}

// ParseCombinedNodeURL parses the combined url of node; it contains,
// - node publish url
// - tls insecure: "#insecure"
// "insecure" fragment will be removed.
func ParseCombinedNodeURL(u *url.URL) (*url.URL, bool, error) {
	if err := IsValidURL(u); err != nil {
		return nil, false, errors.Wrap(err, "invalid combined node url")
	}

	i := NormalizeURL(u)

	insecure := i.Fragment == "insecure"
	i.Fragment = ""

	return i, insecure, nil
}
