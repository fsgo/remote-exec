// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/15

package server

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"path/filepath"

	"github.com/fsgo/fsconf"
	"github.com/fsgo/fsenv"
	"github.com/fsgo/fsgo/fsrpc"

	"github.com/fsgo/remote-exec/internal/packing"
)

var addrDefault = ":8100"

type Config struct {
	Listen string
	Allow  []string
	Users  packing.Users

	CertFile string
	KeyFile  string
}

func (cfg *Config) GetListenAddr() string {
	if cfg.Listen == "" {
		return addrDefault
	}
	return cfg.Listen
}

func (cfg *Config) Merge(c1 *Config) {
	if c1 == nil {
		return
	}
	if cfg.Listen == "" {
		cfg.Listen = c1.Listen
	}
	cfg.Allow = append(cfg.Allow, c1.Allow...)
	cfg.Users = append(cfg.Users, c1.Users...)
	cfg.CertFile = c1.CertFile
	cfg.KeyFile = c1.KeyFile
}

func (cfg *Config) RegisterFlag() {
	flag.StringVar(&cfg.Listen, "l", "", "server listen addr")
}

func (cfg *Config) CmdAllow(name string) bool {
	return true
}

var configPath = flag.String("conf", "./conf/app.toml", "server config")

func parserCnf() (*Config, error) {
	if !fsconf.Exists(*configPath) {
		return &Config{}, nil
	}
	log.Println("loading config:", *configPath)
	abp, err := filepath.Abs(*configPath)
	if err != nil {
		return nil, err
	}
	cr := filepath.Dir(abp)
	fsenv.SetConfRootDir(cr)
	fsenv.SetRootDir(filepath.Dir(cr))
	c := &Config{}
	err = fsconf.Parse(*configPath, c)
	return c, err
}

func (cfg *Config) Run(ser *fsrpc.Server) error {
	addr := cfg.GetListenAddr()
	logData := map[string]any{
		"Listen": addr,
		"Users":  len(cfg.Users),
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		logData["CertFile"] = cfg.CertFile
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return err
		}
		l = tls.NewListener(l, tlsConfig)
	}
	log.Println("server start with", logData)
	return ser.Serve(l)
}
