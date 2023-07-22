// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/16

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/fsgo/fsconf"
)

type Config struct {
	Hosts []*Host
}

func (c *Config) Find(name string) *Host {
	for _, h := range c.Hosts {
		if h.Name == name {
			return h
		}
	}
	return nil
}

type Host struct {
	Name     string
	Address  string
	CertFile string
	UserName string
	Password string
}

var addrDefault = "127.0.0.1:8100"

func (cfg *Host) GetServerAddr() string {
	if cfg.Address == "" {
		return addrDefault
	}
	return cfg.Address
}

func (cfg *Host) merge(b *Host) {
	if b.Address != "" {
		cfg.Address = b.Address
	}
	if b.UserName != "" {
		cfg.UserName = b.UserName
	}
	if b.Password != "" {
		cfg.Password = b.Password
	}
	if b.CertFile != "" {
		cfg.CertFile = b.CertFile
	}
}

func (cfg *Host) Dial(timeout time.Duration) (net.Conn, error) {
	if debug {
		log.Println("cfg.CertFile=", cfg.CertFile)
	}
	address := cfg.GetServerAddr()
	var tlsConfig *tls.Config
	if cfg.CertFile != "" {
		cert, err := os.ReadFile(cfg.CertFile)
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cert)
		tlsConfig = &tls.Config{
			RootCAs:            certPool,
			InsecureSkipVerify: true,
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if tlsConfig == nil || err != nil {
		return conn, err
	}
	tc := tls.Client(conn, tlsConfig)
	if err = tc.HandshakeContext(ctx); err != nil {
		_ = tc.Close()
		return nil, err
	}
	return tc, nil
}

func (cfg *Host) RegisterFlag() {
	flag.StringVar(&cfg.Address, "s", "", "server host:port")
	flag.StringVar(&cfg.UserName, "u", "", "user name")
	flag.StringVar(&cfg.Password, "p", "", "password")
	flag.StringVar(&cfg.CertFile, "cf", "", "cert file")
}

func envValue(ek string, def string) string {
	val := os.Getenv(ek)
	if val != "" {
		return val
	}
	return def
}

var config = &Host{}

var confPath string
var hostName string
var debug bool

func init() {
	config.RegisterFlag()
	hd, err := os.UserHomeDir()
	var confPathDefault string
	if err == nil {
		confPathDefault = filepath.Join(hd, ".config", "remote-exec", "client.toml")
	}
	flag.StringVar(&confPath, "conf", confPathDefault, "config file path")
	stringVar(&hostName, "h", "FS_RE_Host", "", "host alias in config file")
	flag.BoolVar(&debug, "debug", false, "is debugging")
}

func stringVar(p *string, name string, ek string, def string, usage string) {
	usage += "\nwith 'export " + ek + "=xxx' to set the default value"
	flag.StringVar(p, name, envValue(ek, def), usage)
}

func parseFlag() {
	flag.Parse()
	if !fsconf.Exists(confPath) {
		if debug {
			log.Printf("-conf %q not exists", confPath)
		}
		return
	}
	f := &Config{}
	if err := fsconf.Parse(confPath, f); err != nil {
		log.Fatalln(err)
	}

	if len(hostName) > 0 {
		cfg := f.Find(hostName)
		if cfg == nil {
			log.Fatalf("host %q not exists\n", hostName)
		}
		cfg.merge(config)
		config = cfg
		return
	}

	if len(f.Hosts) == 0 {
		return
	}

	if config.Address == "" {
		config = f.Hosts[0]
	}
}
