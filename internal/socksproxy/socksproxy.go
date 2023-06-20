/*
 * Copyright 2023 Petr Svoboda
 */

package socksproxy

import (
	"context"
	"github.com/psvo/flexi-proxy/internal/environment"
	"github.com/psvo/flexi-proxy/internal/proxy"
	"github.com/things-go/go-socks5"
	"github.com/things-go/go-socks5/statute"
	"net"
)

type ctxRequestKey struct{}
type ctxDialerKey struct{}

type myLogger struct {
	env *environment.Environment
}

func (sl *myLogger) Errorf(format string, args ...interface{}) {
	sl.env.Error(format, args...)
}

type myResolver struct {
	env         *environment.Environment
	netResolver net.Resolver
}

func (r *myResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	ips, err := r.netResolver.LookupIP(ctx, "ip", name)
	if err != nil {
		return ctx, nil, err
	}
	ip := ips[0]
	r.env.Debug("resolve: %s => %s\n", name, ip.String())
	return ctx, ip, nil
}

type myRewriter struct {
	env *environment.Environment
}

func (r *myRewriter) Rewrite(ctx context.Context, request *socks5.Request) (context.Context, *statute.AddrSpec) {
	dest := request.DestAddr
	dialer := proxy.ResolveDialer(r.env, dest.FQDN, dest.IP)
	ctx = context.WithValue(ctx, ctxRequestKey{}, request)
	ctx = context.WithValue(ctx, ctxDialerKey{}, dialer)
	r.env.Debug("rewrite: %s => %s", dest.Address(), dialer)
	return ctx, dest
}

type myDialer struct {
	env *environment.Environment
}

func (d *myDialer) dial(ctx context.Context, network, addr string) (conn net.Conn, err error) {
	timeout := d.env.Config().ConnectTimeout()
	req := ctx.Value(ctxRequestKey{}).(*socks5.Request)
	dialer := ctx.Value(ctxDialerKey{}).(proxy.Dialer)
	d.env.Info("CONNECT %s (%s) => %s", req.DestAddr.FQDN, addr, dialer)
	ctx2, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err = dialer.Dial(ctx2, network, addr)
	return conn, err
}

func ListenAndServe(env *environment.Environment) error {
	addr := env.Config().SocksListenAddr
	server := socks5.NewServer(
		socks5.WithLogger(&myLogger{env: env}),
		socks5.WithResolver(&myResolver{env: env}),
		socks5.WithRewriter(&myRewriter{env: env}),
		socks5.WithDial((&myDialer{env: env}).dial),
	)
	env.Info("listening on: socks5://%s", addr)
	return server.ListenAndServe("tcp", addr)
}
