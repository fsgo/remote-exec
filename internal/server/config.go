// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/15

package server

import "flag"

var addrDefault = ":8100"

type Config struct {
	Addr  string
	Allow []string
}

func (cfg *Config) GetAddr() string {
	if cfg.Addr == "" {
		return addrDefault
	}
	return cfg.Addr
}

func (cfg *Config) RegisterFlag() {
	flag.StringVar(&cfg.Addr, "addr", addrDefault, "server listen addr")
}

func (cfg *Config) CmdAllow(name string) bool {
	return true
}

var config = &Config{}

func init() {
	config.RegisterFlag()
}
