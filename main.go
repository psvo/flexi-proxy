/*
 * Copyright 2023 Petr Svoboda
 */

package main

import (
	"flag"
	"github.com/psvo/flexi-proxy/internal/httpproxy"
	"github.com/psvo/flexi-proxy/internal/proxy"
	"github.com/psvo/flexi-proxy/internal/socksproxy"
	"log"
	"os"
	"sync"
	"time"
)

func runHttpProxy(loader *proxy.EnvLoader, logPrefix string) {
	env := loader.Env().WithLogger(mkLogger(logPrefix))
	if env.Config().HttpListenAddr == "" {
		return
	}
	if err := httpproxy.ListenAndServe(env); err != nil {
		panic(err)
	}
}

func runSocksProxy(loader *proxy.EnvLoader, logPrefix string) {
	env := loader.Env().WithLogger(mkLogger(logPrefix))
	if env.Config().SocksListenAddr == "" {
		return
	}
	if err := socksproxy.ListenAndServe(env); err != nil {
		panic(err)
	}
}

func runAsync(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

func mkLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix+" ", log.LstdFlags|log.Lmsgprefix)
}

func main() {
	var confFilePath = "proxy.toml"
	flag.StringVar(&confFilePath, "c", confFilePath, "path to configuration file")
	flag.Parse()
	loader, err := proxy.NewEnvironmentLoader(confFilePath, 3*time.Second, mkLogger("config"))
	if err != nil {
		panic(err)
	}
	defer loader.Stop()
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	runAsync(wg, func() {
		runHttpProxy(loader, "http")
	})
	runAsync(wg, func() {
		runSocksProxy(loader, "socks")
	})
}
