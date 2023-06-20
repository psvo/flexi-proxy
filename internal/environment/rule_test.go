/*
 * Copyright 2023 Petr Svoboda
 */

package environment

import (
	"net"
	"testing"
)

func TestDomainMatcher_Exact(t *testing.T) {
	pat := "the.test"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	dom := "the.test"
	if !m.Matches(dom, nil) {
		t.Fatalf("Domain matcher on `%s` should match `%s`", pat, dom)
	}
}

func TestDomainMatcher_NoSubdomain(t *testing.T) {
	pat := "the.test"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	dom := "sub.the.test"
	if m.Matches(dom, nil) {
		t.Fatalf("Domain matcher on `%s` should not match subdomain `%s`", pat, dom)
	}
}

func TestDomainMatcher_NoParent(t *testing.T) {
	pat := "sub.the.test"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	dom := "the.test"
	if m.Matches(dom, nil) {
		t.Fatalf("Domain matcher on `%s` should not match parent domain `%s`", pat, dom)
	}
}

func TestSubdomainMatcher_Exact(t *testing.T) {
	pat := ".the.test"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	dom := "the.test"
	if !m.Matches(dom, nil) {
		t.Fatalf("Subdomain matcher on `%s` should match `%s`", pat, dom)
	}
}

func TestSubdomainMatcher_Subdomain(t *testing.T) {
	pat := ".the.test"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	dom := "sub.the.test"
	if !m.Matches(dom, nil) {
		t.Fatalf("Subdomain matcher on `%s` should match `%s`", pat, dom)
	}
}

func TestSubdomainMatcher_SubdomainDot(t *testing.T) {
	pat := "."
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	dom := "sub.the.test"
	if !m.Matches(dom, nil) {
		t.Fatalf("Subdomain matcher on `%s` should match `%s`", pat, dom)
	}
}

func TestSubdomainMatcher_SubdomainDotAndNoHost(t *testing.T) {
	pat := "."
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	dom := ""
	if m.Matches(dom, nil) {
		t.Fatalf("Subdomain matcher on `%s` should not match `%s`", pat, dom)
	}
}

func TestSubdomainMatcher_NoParent(t *testing.T) {
	pat := ".sub.the.test"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	dom := "the.test"
	if m.Matches(dom, nil) {
		t.Fatalf("Subdomain matcher on `%s` should not match parent domain `%s`", pat, dom)
	}
}

func TestCidrMatcher_MatchesIPv4(t *testing.T) {
	pat := "192.168.5.1/24"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	ip := net.ParseIP("192.168.5.10")
	if ip == nil || !m.Matches("", ip) {
		t.Fatalf("CIDR matcher on `%s` should match IP `%v`", pat, ip)
	}
}

func TestCidrMatcher_NotMatchesIPv4(t *testing.T) {
	pat := "192.168.5.1/24"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	ip := net.ParseIP("192.168.6.10")
	if ip == nil || m.Matches("", ip) {
		t.Fatalf("CIDR matcher on `%s` should not match IP `%v`", pat, ip)
	}
}

func TestNewMatcher_NoIPv4(t *testing.T) {
	pat := "192.168.5.1"
	_, err := newMatcher(pat)
	if err == nil {
		t.Fatalf("Pattern `%s` should be rejected", pat)
	}
}

func TestNewMatcher_BadIPv4CIDR(t *testing.T) {
	pat := "192.168.5.1/"
	_, err := newMatcher(pat)
	if err == nil {
		t.Fatalf("Pattern `%s` should be rejected", pat)
	}
}

func TestCidrMatcher_MatchesIPv6(t *testing.T) {
	pat := "1::/64"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	ip := net.ParseIP("1::10")
	if ip == nil || !m.Matches("", ip) {
		t.Fatalf("CIDR matcher on `%s` should match IP `%v`", pat, ip)
	}
}

func TestCidrMatcher_NotMatchesIPv6(t *testing.T) {
	pat := "1::/64"
	m, err := newMatcher(pat)
	if err != nil {
		t.Fatalf("Failed to create matcher on pattern `%s`: %v", pat, err)
	}
	ip := net.ParseIP("2::10")
	if ip == nil || m.Matches("", ip) {
		t.Fatalf("CIDR matcher on `%s` should not match IP `%v`", pat, ip)
	}
}

func TestNewMatcher_NoIPv6(t *testing.T) {
	pat := "1::"
	_, err := newMatcher(pat)
	if err == nil {
		t.Fatalf("Pattern `%s` should be rejected", pat)
	}
}

func TestNewMatcher_BadIPv6CIDR(t *testing.T) {
	pat := "1::/"
	_, err := newMatcher(pat)
	if err == nil {
		t.Fatalf("Pattern `%s` should be rejected", pat)
	}
}

func TestNewMatcher_Empty(t *testing.T) {
	pat := ""
	_, err := newMatcher(pat)
	if err == nil {
		t.Fatalf("Pattern `%s` should be rejected", pat)
	}
}
