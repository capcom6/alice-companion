package config

import (
	"github.com/capcom6/alice-companion/internal/alice"
	"github.com/capcom6/alice-companion/internal/example"
	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/fiberfx/openapi"
	"github.com/go-core-fx/httpfx"
	"github.com/go-core-fx/sqlfx"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"config",
		fx.Provide(New, fx.Private),
		fx.Provide(
			func(cfg Config) fiberfx.Config {
				return fiberfx.Config{
					Address:     cfg.HTTP.Address,
					ProxyHeader: cfg.HTTP.ProxyHeader,
					Proxies:     cfg.HTTP.Proxies,
				}
			},
			func(cfg Config) openapi.Config {
				return openapi.Config{
					Enabled:    cfg.HTTP.OpenAPI.Enabled,
					PublicHost: cfg.HTTP.OpenAPI.PublicHost,
					PublicPath: cfg.HTTP.OpenAPI.PublicPath,
				}
			},
			func(cfg Config) httpfx.Config {
				return httpfx.Config{
					ProxyURL:            cfg.HTTPClient.ProxyURL,
					Bypass:              cfg.HTTPClient.Bypass,
					Timeout:             cfg.HTTPClient.Timeout,
					MaxIdleConns:        cfg.HTTPClient.MaxIdleConns,
					MaxIdleConnsPerHost: cfg.HTTPClient.MaxIdleConnsPerHost,
					IdleConnTimeout:     cfg.HTTPClient.IdleConnTimeout,
					TLS: httpfx.TLSConfig{
						RootCAFile:          cfg.HTTPClient.TLS.RootCAFile,
						RootCAPEM:           cfg.HTTPClient.TLS.RootCAPEM,
						RootCAReplaceSystem: cfg.HTTPClient.TLS.RootCAReplaceSystem,
					},
				}
			},
			func(cfg Config) sqlfx.Config {
				return sqlfx.Config{
					URL:             cfg.Database.URL,
					ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
					ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
					MaxOpenConns:    cfg.Database.MaxOpenConns,
					MaxIdleConns:    cfg.Database.MaxIdleConns,
				}
			},
			func(cfg Config) cachefx.Config {
				return cachefx.Config{
					URL:           cfg.Cache.URL,
					SweepInterval: cfg.Cache.SweepInterval,
				}
			},
		),
		fx.Provide(func(cfg Config) alice.Config {
			return alice.Config{
				SharedSecret: cfg.Alice.SharedSecret,
			}
		}),
		fx.Provide(func(cfg Config) example.Config {
			return example.Config{
				Example: cfg.Example.Example,
			}
		}),
	)
}
