/*
 * Copyright 2023 Petr Svoboda
 */

package proxy

import (
	"context"
	"fmt"
	"github.com/psvo/flexi-proxy/internal/environment"
	"net"
)

type dialerDirect struct {
	env  *environment.Environment
	dial dialerFunc
}

func (d *dialerDirect) String() string {
	return "DIRECT"
}

func (d *dialerDirect) Dial(ctx context.Context, network, address string) (conn net.Conn, err error) {
	defer func() {
		if err != nil && conn != nil {
			_ = conn.Close()
		}
	}()
	d.env.Debug("dial connecting: %s => %s://%s", d, network, address)

	ctx, cancel := context.WithTimeout(ctx, d.env.Config().ConnectTimeout())
	defer cancel()

	conn, err = d.dial(ctx, network, address)
	if err != nil {
		return nil, d.mkError(err, "unable to connect")
	}

	d.env.Debug("dial: connected: %s => %s://%s", d, network, address)
	return conn, nil
}

func (d *dialerDirect) mkError(cause error, format string, args ...interface{}) error {
	return fmt.Errorf("%s: %s: %w",
		d, fmt.Sprintf(format, args...), cause,
	)
}
