package digest

import (
	"context"
	"fmt"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"net/http"
	"strings"
	"time"

	"github.com/ProtoconNet/mitum-currency/v3/digest/network"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sync/singleflight"
)

var (
	HTTP2EncoderHintHeader = http.CanonicalHeaderKey("x-mitum-encoder-hint")
	HALMimetype            = "application/hal+json; charset=utf-8"
)

var (
	HandlerPathNodeInfo                   = `/`
	HandlerPathCurrencies                 = `/currency`
	HandlerPathCurrency                   = `/currency/{currencyid:.*}`
	HandlerPathManifests                  = `/block/manifests`
	HandlerPathOperations                 = `/block/operations`
	HandlerPathOperation                  = `/block/operation/{hash:(?i)[0-9a-z][0-9a-z]+}`
	HandlerPathBlockByHeight              = `/block/{height:[0-9]+}`
	HandlerPathBlockByHash                = `/block/{hash:(?i)[0-9a-z][0-9a-z]+}`
	HandlerPathOperationsByHeight         = `/block/{height:[0-9]+}/operations`
	HandlerPathManifestByHeight           = `/block/{height:[0-9]+}/manifest`
	HandlerPathManifestByHash             = `/block/{hash:(?i)[0-9a-z][0-9a-z]+}/manifest`
	HandlerPathAccount                    = `/account/{address:(?i)` + base.REStringAddressString + `}`            // revive:disable-line:line-length-limit
	HandlerPathAccountOperations          = `/account/{address:(?i)` + base.REStringAddressString + `}/operations` // revive:disable-line:line-length-limit
	HandlerPathAccounts                   = `/accounts`
	HandlerPathOperationBuildFactTemplate = `/builder/operation/fact/template/{fact:[\w][\w\-]*}`
	HandlerPathOperationBuildFact         = `/builder/operation/fact`
	HandlerPathOperationBuildSign         = `/builder/operation/sign`
	HandlerPathOperationBuild             = `/builder/operation`
	HandlerPathSend                       = `/builder/send`
	HandlerPathQueueSend                  = `/builder/send/queue`
	HandelrPathEventOperation             = `/event/operation/{hash:(?i)[0-9a-z][0-9a-z]+}`
	HandelrPathEventAccount               = `/event/account/{address:(?i)` + base.REStringAddressString + `}`
	HandlerPathEventContract              = `/event/contract/{address:(?i)` + base.REStringAddressString + `}`
)

var (
	UnknownProblem     = NewProblem(DefaultProblemType, "unknown problem occurred")
	UnknownProblemJSON []byte
)

var GlobalItemsLimit int64 = 10

func init() {
	if b, err := JSON.Marshal(UnknownProblem); err != nil {
		panic(err)
	} else {
		UnknownProblemJSON = b
	}
}

type Handlers struct {
	*zerolog.Logger
	networkID       base.NetworkID
	encs            *encoder.Encoders
	enc             encoder.Encoder
	database        *Database
	cache           Cache
	queue           chan RequestWrapper
	nodeInfoHandler NodeInfoHandler
	send            func(interface{}) (base.Operation, error)
	client          func() (*isaacnetwork.BaseClient, *quicmemberlist.Memberlist, error)
	router          *mux.Router
	routes          map[ /* path */ string]*mux.Route
	itemsLimiter    func(string /* request type */) int64
	rg              *singleflight.Group
	expireNotFilled time.Duration
}

func NewHandlers(
	ctx context.Context,
	networkID base.NetworkID,
	encs *encoder.Encoders,
	enc encoder.Encoder,
	st *Database,
	cache Cache,
	router *mux.Router,
	queue chan RequestWrapper,
) *Handlers {
	var log *logging.Logging
	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return nil
	}

	return &Handlers{
		Logger:          log.Log(),
		networkID:       networkID,
		encs:            encs,
		enc:             enc,
		database:        st,
		cache:           cache,
		queue:           queue,
		router:          router,
		routes:          map[string]*mux.Route{},
		itemsLimiter:    DefaultItemsLimiter,
		rg:              &singleflight.Group{},
		expireNotFilled: time.Second * 3,
	}
}

func (hd *Handlers) Initialize() error {
	cors := handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
	hd.router.Use(cors)

	hd.setHandlers()

	return nil
}

func (hd *Handlers) SetLimiter(f func(string) int64) *Handlers {
	hd.itemsLimiter = f

	return hd
}

func (hd *Handlers) Cache() Cache {
	return hd.cache
}

func (hd *Handlers) Router() *mux.Router {
	return hd.router
}

func (hd *Handlers) Routes() map[string]*mux.Route {
	return hd.routes
}

func (hd *Handlers) Handler() http.Handler {
	return network.HTTPLogHandler(hd.router, hd.Logger)
}

func (hd *Handlers) setHandlers() {
	_ = hd.setHandler(HandlerPathCurrencies, hd.handleCurrencies, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathCurrency, hd.handleCurrency, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathManifests, hd.handleManifests, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathOperations, hd.handleOperations, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathOperation, hd.handleOperation, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathOperationsByHeight, hd.handleOperationsByHeight, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathManifestByHeight, hd.handleManifestByHeight, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathManifestByHash, hd.handleManifestByHash, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathBlockByHeight, hd.handleBlock, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathBlockByHash, hd.handleBlock, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathAccount, hd.handleAccount, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathAccountOperations, hd.handleAccountOperations, true).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathAccounts, hd.handleAccounts, true).
		Methods(http.MethodOptions, "GET")
	// _ = hd.setHandler(HandlerPathOperationBuildFactTemplate, hd.handleOperationBuildFactTemplate, true).
	// 	Methods(http.MethodOptions, "GET")
	// _ = hd.setHandler(HandlerPathOperationBuildFact, hd.handleOperationBuildFact, false).
	// 	Methods(http.MethodOptions, http.MethodPost)
	// _ = hd.setHandler(HandlerPathOperationBuildSign, hd.handleOperationBuildSign, false).
	// 	Methods(http.MethodOptions, http.MethodPost)
	// _ = hd.setHandler(HandlerPathOperationBuild, hd.handleOperationBuild, true).
	// 	Methods(http.MethodOptions, http.MethodGet, http.MethodPost)
	_ = hd.setHandler(HandlerPathSend, hd.handleSend, false).
		Methods(http.MethodOptions, http.MethodPost)
	_ = hd.setHandler(HandlerPathQueueSend, hd.handleQueueSend, false).
		Methods(http.MethodOptions, http.MethodPost)
	_ = hd.setHandler(HandlerPathNodeInfo, hd.handleNodeInfo, true).
		Methods(http.MethodOptions, "GET")
}

func (hd *Handlers) setHandler(prefix string, h network.HTTPHandlerFunc, useCache bool) *mux.Route {
	var handler http.Handler
	if !useCache {
		handler = http.HandlerFunc(h)
	} else {
		ch := NewCachedHTTPHandler(hd.cache, h)

		handler = ch
	}

	var name string
	if prefix == "" || prefix == "/" {
		name = "root"
	} else {
		name = prefix
	}

	var route *mux.Route
	if r := hd.router.Get(name); r != nil {
		route = r
	} else {
		route = hd.router.Name(name)
	}

	/*
		if rules, found := hd.rateLimit[prefix]; found {
			handler = process.NewRateLimitMiddleware(
				process.NewRateLimit(rules, limiter.Rate{Limit: -1}), // NOTE by default, unlimited
				hd.rateLimitStore,
			).Middleware(handler)

			hd.Log().Debug().Str("prefix", prefix).Msg("ratelimit middleware attached")
		}
	*/

	route = route.
		Path(prefix).
		Handler(handler)

	hd.routes[prefix] = route

	return route
}

func (hd *Handlers) combineURL(path string, pairs ...string) (string, error) {
	if n := len(pairs); n%2 != 0 {
		return "", errors.Errorf("combine url; uneven pairs to combine url")
	} else if n < 1 {
		u, err := hd.routes[path].URL()
		if err != nil {
			return "", errors.Wrap(err, "combine url")
		}
		return u.String(), nil
	}

	u, err := hd.routes[path].URLPath(pairs...)
	if err != nil {
		return "", errors.Wrap(err, "combine url")
	}
	return u.String(), nil
}

func CacheKeyPath(r *http.Request) string {
	return r.URL.Path
}

func CacheKey(key string, s ...string) string {
	var l []string
	var notempty bool
	for i := len(s) - 1; i >= 0; i-- {
		a := s[i]

		if !notempty {
			if len(strings.TrimSpace(a)) < 1 {
				continue
			}
			notempty = true
		}

		l = append(l, a)
	}

	r := make([]string, len(l))
	for i := range l {
		r[len(l)-1-i] = l[i]
	}

	return fmt.Sprintf("%s-%s", key, strings.Join(r, ","))
}

func DefaultItemsLimiter(string) int64 {
	return GlobalItemsLimit
}
