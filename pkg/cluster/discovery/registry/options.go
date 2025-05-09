package registry

import (
	"context"
	"crypto/tls"
	"time"
)

type Options struct {
	// Other options for implementations of the interface
	// can be stored in a context
	Context   context.Context
	TLSConfig *tls.Config
	Addrs     []string
	Timeout   time.Duration
	Secure    bool
}

type RegisterOptions struct {
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
	TTL     time.Duration
}

type WatchOptions struct {
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
	// Specify a service to watch
	// If blank, the watch is for all services
	Service string
}

type DeregisterOptions struct {
	Context context.Context
}

type GetOptions struct {
	Context context.Context
}

type ListOptions struct {
	Context context.Context
}

func NewOptions(opts ...Option) *Options {
	options := Options{
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &options
}

// Addrs is the registry addresses to use.
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

func Timeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// Secure communication with the registry.
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// Specify TLS Config.
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

func RegisterTTL(t time.Duration) RegisterOption {
	return func(o *RegisterOptions) {
		o.TTL = t
	}
}

func RegisterContext(ctx context.Context) RegisterOption {
	return func(o *RegisterOptions) {
		o.Context = ctx
	}
}

// Watch a service.
func WatchService(name string) WatchOption {
	return func(o *WatchOptions) {
		o.Service = name
	}
}

func WatchContext(ctx context.Context) WatchOption {
	return func(o *WatchOptions) {
		o.Context = ctx
	}
}

func DeregisterContext(ctx context.Context) DeregisterOption {
	return func(o *DeregisterOptions) {
		o.Context = ctx
	}
}

func GetContext(ctx context.Context) GetOption {
	return func(o *GetOptions) {
		o.Context = ctx
	}
}

func ListContext(ctx context.Context) ListOption {
	return func(o *ListOptions) {
		o.Context = ctx
	}
}

type servicesKey struct{}

// Services is an option that preloads service data.
func Services(s map[string][]*Service) Option {
	return func(o *Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, servicesKey{}, s)
	}
}
