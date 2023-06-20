/*
 * Copyright 2023 Petr Svoboda
 */

package proxy

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/psvo/flexi-proxy/internal/environment"
	"log"
	"os"
	"time"
)

type EnvLoader struct {
	configFilePath string
	pollPeriod     time.Duration
	env            *environment.Environment
	ctx            context.Context
	cancel         context.CancelFunc
	lastStat       os.FileInfo
}

func NewEnvironmentLoader(configFilePath string, pollPeriod time.Duration, logger *log.Logger) (*EnvLoader, error) {
	env := environment.NewEnvironment(logger)
	stat, err := os.Stat(configFilePath)
	if err != nil {
		return nil, err
	}
	if err := loadConfig(env, configFilePath); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	envLoader := &EnvLoader{
		configFilePath: configFilePath,
		pollPeriod:     pollPeriod,
		env:            env,
		ctx:            ctx,
		cancel:         cancel,
		lastStat:       stat,
	}
	go envLoader.runWatcher()
	return envLoader, nil
}

func loadConfig(env *environment.Environment, configFilePath string) error {
	cfg := &environment.Config{
		ConnectTimeoutMillis: 10_000,
		HttpListenAddr:       "127.0.0.1:8001",
		SocksListenAddr:      "127.0.0.1:8002",
	}
	meta, err := toml.DecodeFile(configFilePath, cfg)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %s: %w", configFilePath, err)
	}
	if uKeys := meta.Undecoded(); len(uKeys) > 0 {
		env.Warn("Config file has unknown fields: %v", uKeys)
	}
	env.Debug("Loaded configuration: %+v", cfg)
	err = env.SetConfig(cfg)
	if err != nil {
		return fmt.Errorf("cannot use the loaded configuration: %w", err)
	}
	return nil
}

func (l *EnvLoader) Env() *environment.Environment {
	return l.env
}

func (l *EnvLoader) Stop() {
	l.cancel()
}

func (l *EnvLoader) runWatcher() {
	env := l.env
	for {
		select {
		case <-l.ctx.Done():
			return
		case <-time.After(l.pollPeriod):
			//noop
		}
		env.Debug("Polling for changes")
		stat, err := os.Stat(l.configFilePath)
		if err != nil {
			env.Warn("Cannot stat config file: %v", err)
		} else if l.lastStat == nil || stat.Size() != l.lastStat.Size() || stat.ModTime() != l.lastStat.ModTime() {
			env.Info("Detected changes in config file `%s`, reloading", l.configFilePath)
			if err := loadConfig(l.env, l.configFilePath); err != nil {
				env.Warn("Cannot load config file: %v", err)
			}
		}
		l.lastStat = stat
	}
}
