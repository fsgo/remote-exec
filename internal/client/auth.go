// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/16

package client

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/fsgo/fsgo/fsrpc"

	"github.com/fsgo/remote-exec/internal/packing"
)

func sendAuth(rw fsrpc.RequestWriter) {
	if config.UserName == "" {
		return
	}
	user := &packing.User{
		Name: strings.TrimSpace(config.UserName),
		Psw:  strings.TrimSpace(config.Password),
	}
	ah := &fsrpc.AuthHandler{
		ClientData: func(ctx context.Context) *fsrpc.AuthData {
			return user.ToAuthData()
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ret := ah.Client(ctx, rw)
	if debug && ret == nil {
		log.Println("send auth success")
	}
	if ret != nil {
		log.Fatalln(ret)
	}
}
