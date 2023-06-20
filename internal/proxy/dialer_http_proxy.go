/*
 * Copyright 2023 Petr Svoboda
 */

package proxy

import (
	"bufio"
	"context"
	"fmt"
	"github.com/psvo/flexi-proxy/internal/environment"
	"net"
	"net/http"
)

type dialerHttpProxy struct {
	env       *environment.Environment
	dial      dialerFunc
	proxyAddr string
	fqdn      string
}

func (d *dialerHttpProxy) String() string {
	return "PROXY http://" + d.proxyAddr
}

func (d *dialerHttpProxy) Dial(ctx context.Context, network, address string) (conn net.Conn, err error) {
	defer func() {
		if err != nil && conn != nil {
			_ = conn.Close()
		}
	}()
	conn, err = d.dial(ctx, network, d.proxyAddr)
	if err != nil {
		return nil, d.mkError(err, "unable to connect")
	}

	_, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\n\r\n", address)
	if err != nil {
		return nil, d.mkError(err, "unable send request")
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, d.mkError(err, "unable to read response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, d.mkError(fmt.Errorf("%s", resp.Status), "got response status")
	}

	return conn, nil
}

func (d *dialerHttpProxy) mkError(cause error, format string, args ...interface{}) error {
	return fmt.Errorf("%s: %s: %w",
		d, fmt.Sprintf(format, args...), cause,
	)
}
