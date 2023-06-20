/*
 * Copyright 2023 Petr Svoboda
 */

package httpproxy

import (
	"context"
	"fmt"
	"github.com/psvo/flexi-proxy/internal/environment"
	"github.com/psvo/flexi-proxy/internal/proxy"
	"github.com/things-go/go-socks5/bufferpool"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
)

type myHandler struct {
	env        *environment.Environment
	bufferPool bufferpool.BufPool
	resolver   net.Resolver
}

func (h *myHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer func() {
		_ = req.Body.Close()
	}()
	var err error
	if req.Method == http.MethodConnect {
		h.handleConnectRequest(res, req)
	} else {
		h.handleHttpRequest(res, req)
	}
	if err != nil {
		h.env.Error("bad request: %s", err)
		res.WriteHeader(http.StatusBadGateway)
	}
}

func (h *myHandler) handleConnectRequest(res http.ResponseWriter, req *http.Request) {
	addr := req.RequestURI
	dialer := h.resolveDialer(addr)
	h.env.Info("%s %s => %s",
		req.Method, req.RequestURI, dialer,
	)
	targetConn, err := dialer.Dial(req.Context(), "tcp", addr)
	if err != nil {
		h.env.Error("%s %s => %s", req.Method, req.RequestURI, err)
		res.WriteHeader(http.StatusBadGateway)
		return
	}
	defer func() { _ = targetConn.Close() }()
	clientConn, rw, err := (res.(http.Hijacker)).Hijack()
	if err != nil {
		h.logError(dialer, req.Method, req.RequestURI, fmt.Errorf("forwarding setup failed"))
		res.WriteHeader(http.StatusBadGateway)
		return
	}
	defer func() { _ = clientConn.Close() }()
	_, err = rw.WriteString("HTTP/1.1 200 Connection established\r\n\r\n")
	if err != nil {
		h.logError(dialer, req.Method, req.RequestURI, fmt.Errorf("writing response failed"))
	}
	err = rw.Flush()
	if err != nil {
		h.logError(dialer, req.Method, req.RequestURI, fmt.Errorf("flushing response failed"))
	}

	doProxy := func(wg *sync.WaitGroup, dst io.Writer, src io.Reader) {
		defer wg.Done()
		err := h.proxy(dst, src)
		if err != nil {
			h.logError(dialer, req.Method, req.RequestURI, fmt.Errorf("proxy tunnel error: %w", err))
		}
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go doProxy(wg, targetConn, rw)
	go doProxy(wg, clientConn, targetConn)
	wg.Wait()
}

func (h *myHandler) logError(dialer proxy.Dialer, method, uri string, err error) {
	h.env.Error("%s %s => %s: %v", method, uri, dialer, err)
}

func (h *myHandler) handleHttpRequest(res http.ResponseWriter, req *http.Request) {
	dialer := h.resolveDialer(req.URL.Host)
	h.env.Info("%s %s => %s", req.Method, req.URL, dialer)
	rp := httputil.ReverseProxy{
		Rewrite:  func(*httputil.ProxyRequest) { /* noop */ },
		ErrorLog: h.env.Logger(),
		Transport: &http.Transport{
			Proxy:       nil,
			DialContext: dialer.Dial,
		},
	}
	rp.ServeHTTP(res, req)
}

func (h *myHandler) resolveDialer(addr string) proxy.Dialer {
	parts := strings.SplitN(addr, ":", 2)
	host := parts[0]
	ip := net.ParseIP(host)
	if ip == nil {
		ctx, cancel := context.WithTimeout(context.Background(), h.env.Config().ConnectTimeout())
		defer cancel()
		ips, err := h.resolver.LookupIP(ctx, "ip", host)
		if err != nil {
			h.env.Warn("IP lookup failed: %v", err)
		} else {
			ip = ips[0]
		}
	} else {
		host = ""
	}
	return proxy.ResolveDialer(h.env, host, ip)
}

type closeWriter interface {
	CloseWrite() error
}

func (h *myHandler) proxy(dst io.Writer, src io.Reader) error {
	buf := h.bufferPool.Get()
	defer h.bufferPool.Put(buf)
	_, err := io.CopyBuffer(dst, src, buf[:cap(buf)])
	if tcpConn, ok := dst.(closeWriter); ok {
		_ = tcpConn.CloseWrite()
	}
	return err
}

func ListenAndServe(env *environment.Environment) error {
	cfg := env.Config()
	addr := cfg.HttpListenAddr
	server := &http.Server{
		Addr: addr,
		Handler: &myHandler{
			env:        env,
			bufferPool: bufferpool.NewPool(32 * 1024),
		},
		ReadTimeout:    cfg.ReadTimeout(),
		WriteTimeout:   cfg.WriteTimeout(),
		MaxHeaderBytes: 16 * 1024,
		ErrorLog:       env.Logger(),
	}
	env.Info("listening on: http://%s", addr)
	return server.ListenAndServe()
}
