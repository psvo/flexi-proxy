/*
 * Copyright 2023 Petr Svoboda
 */

package environment

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

type verbosity int

const (
	Error verbosity = iota
	Warn
	Info
	Debug
)

func NewEnvironment(logger *log.Logger) *Environment {
	e := &Environment{
		logger: logger,
		config: &atomic.Pointer[Config]{},
	}
	e.config.Store(&Config{})
	return e
}

type Environment struct {
	logger *log.Logger
	config *atomic.Pointer[Config]
}

func (e *Environment) WithLogger(logger *log.Logger) *Environment {
	return &Environment{
		logger: logger,
		config: e.config,
	}
}

func (e *Environment) Config() *Config {
	return e.config.Load()
}

func (e *Environment) Logger() *log.Logger {
	return e.logger
}

func (e *Environment) Error(fmt string, args ...interface{}) {
	if e.Config().Verbosity >= Error {
		e.logger.Printf("ERROR "+fmt, args...)
	}
}

func (e *Environment) Warn(fmt string, args ...interface{}) {
	if e.Config().Verbosity >= Warn {
		e.logger.Printf("WARN "+fmt, args...)
	}
}

func (e *Environment) Info(fmt string, args ...interface{}) {
	if e.Config().Verbosity >= Info {
		e.logger.Printf("INFO "+fmt, args...)
	}
}

func (e *Environment) Debug(fmt string, args ...interface{}) {
	if e.Config().Verbosity >= Debug {
		e.logger.Printf("DEBUG "+fmt, args...)
	}
}

func (e *Environment) SetConfig(cfg *Config) error {
	if cfg.Rules == nil {
		return fmt.Errorf("no rules were defined")
	}
	for i := range cfg.Rules {
		rule := &cfg.Rules[i]
		if rule.Proxy == "" {
			rule.url = url.URL{}
		} else {
			u, err := parseProxyUrl(rule.Proxy)
			if err != nil {
				return fmt.Errorf("rule[%d] proxy `%s` %w", i, rule.Proxy, err)
			}
			rule.url = *u
		}
		matchers, err := buildMatchers(rule.Patterns)
		if err != nil {
			return fmt.Errorf("rule[%d] %w", i, err)
		}
		rule.matchers = matchers
	}
	e.config.Store(cfg)
	return nil
}

type Config struct {
	HttpListenAddr       string
	SocksListenAddr      string
	ConnectTimeoutMillis int
	ReadTimeoutMillis    int
	WriteTimeoutMillis   int
	Verbosity            verbosity
	Rules                []Rule
}

// possible config switching URL: detectportal.firefox.com

func (c *Config) ConnectTimeout() time.Duration {
	return time.Duration(c.ConnectTimeoutMillis) * time.Millisecond
}

func (c *Config) ReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeoutMillis) * time.Millisecond
}

func (c *Config) WriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeoutMillis) * time.Millisecond
}

func (e *Environment) ResolveProxyRule(normalizedDomainName string, ip net.IP) *Rule {
	cfg := e.Config()
	for i := range cfg.Rules {
		rule := cfg.Rules[i]
		for j, m := range rule.matchers {
			if m.Matches(normalizedDomainName, ip) {
				e.Debug("rule[%d]/pattern[%d](%s) matches: %s %v", i, j, rule.Patterns[j], normalizedDomainName, ip)
				return &rule
			} else {
				e.Debug("rule[%d]/pattern[%d](%s) does not match: %s %v", i, j, rule.Patterns[j], normalizedDomainName, ip)
			}
		}
	}
	e.Debug("no pattern matches: %s %v", normalizedDomainName, ip)
	return nil
}
