/*
 * Copyright 2023 Petr Svoboda
 */

package proxy

import (
	"context"
	"fmt"
	"github.com/psvo/flexi-proxy/internal/environment"
	"net"
	"strings"
)

type Dialer interface {
	fmt.Stringer
	Dial(ctx context.Context, network, address string) (net.Conn, error)
}

func ResolveDialer(env *environment.Environment, fqdn string, ip net.IP) Dialer {
	env.Debug("resolve: %v / %v", fqdn, ip)
	normalizedDomainName := strings.ToLower(fqdn)
	rule := env.ResolveProxyRule(normalizedDomainName, ip)
	switch rule.ProxyScheme() {
	case "":
		return &dialerDirect{
			env:  env,
			dial: mkDialerFunc(env),
		}
	case "http":
		return &dialerHttpProxy{
			env:       env,
			dial:      mkDialerFunc(env),
			proxyAddr: rule.ProxyAddr(),
			fqdn:      normalizedDomainName,
		}
	default:
		panic("unknown rule proxy schema: " + rule.ProxyScheme())
	}
}

type dialerFunc func(ctx context.Context, network, address string) (conn net.Conn, err error)

func mkDialerFunc(env *environment.Environment) dialerFunc {
	cfg := env.Config()
	netDialer := net.Dialer{
		Timeout: cfg.ConnectTimeout(),
	}
	return func(ctx context.Context, network, address string) (conn net.Conn, err error) {
		defer func() {
			if err != nil && conn != nil {
				_ = conn.Close()
			}
		}()
		env.Debug("dial: connecting: %s://%s", network, address)

		ctx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout())
		defer cancel()

		conn, err = netDialer.DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}

		env.Debug("dial: connected: %s://%s", network, address)
		return conn, nil
	}
}
