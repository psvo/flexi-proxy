/*
 * Copyright 2023 Petr Svoboda
 */

package environment

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

type Rule struct {
	Proxy    string
	Patterns []string
	url      url.URL
	matchers []matcher
}

func (r *Rule) ProxyScheme() string {
	if r == nil {
		return ""
	}
	return r.url.Scheme
}

func (r *Rule) ProxyAddr() string {
	if r == nil {
		return ""
	}
	return r.url.Host
}

type matcher interface {
	Matches(normalizedDomainName string, ip net.IP) bool
}

type domainMatcher struct {
	domain string
}

func (m *domainMatcher) Matches(normalizedDomainName string, _ net.IP) bool {
	return normalizedDomainName == m.domain
}

type subdomainMatcher struct {
	length int
	domain string
}

func (m *subdomainMatcher) Matches(normalizedDomainName string, _ net.IP) bool {
	length := len(normalizedDomainName)
	if length > 0 {
		if m.length == 0 {
			return true
		} else if length == m.length {
			return normalizedDomainName == m.domain
		} else if length > m.length {
			dot := length - m.length - 1
			return normalizedDomainName[dot] == '.' && normalizedDomainName[dot+1:] == m.domain
		} else {
			return false
		}
	} else {
		return false
	}
}

type cidrMatcher struct {
	cidr *net.IPNet
}

func (m *cidrMatcher) Matches(_ string, ip net.IP) bool {
	return m.cidr.Contains(ip)
}

func newMatcher(pattern string) (matcher, error) {
	if ip := net.ParseIP(pattern); pattern == "" || ip != nil {
		return nil, fmt.Errorf("domain name pattern or CIDR is required")
	}
	if strings.IndexByte(pattern, '/') >= 0 {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return nil, err
		}
		m := cidrMatcher{
			cidr: ipNet,
		}
		return &m, nil
	}
	if pattern[0] == '.' {
		m := subdomainMatcher{
			length: len(pattern) - 1,
			domain: pattern[1:],
		}
		return &m, nil
	}
	m := domainMatcher{
		domain: pattern,
	}
	return &m, nil
}

func parseProxyUrl(proxyUrl string) (*url.URL, error) {
	u, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" {
		return nil, fmt.Errorf("has unsupported scheme `%s`", u.Scheme)
	}
	if u.Path == "/" {
		u.Path = ""
	}
	if u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("has unsupported parts")
	}
	return u, nil
}

func buildMatchers(patterns []string) ([]matcher, error) {
	matchers := make([]matcher, len(patterns))
	for i := range patterns {
		m, err := newMatcher(patterns[i])
		if err != nil {
			return nil, fmt.Errorf("pattern[%d]: %w", i, err)
		}
		matchers[i] = m
	}
	return matchers, nil
}
