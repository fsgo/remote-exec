// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/16

package server

import (
	"context"
	"fmt"
	"log"

	"github.com/fsgo/fsgo/fsrpc"
)

func serverAuth() *fsrpc.AuthHandler {
	if len(config.Users) == 0 {
		return &fsrpc.AuthHandler{}
	}
	return &fsrpc.AuthHandler{
		ServerCheck: func(ctx context.Context, ar *fsrpc.AuthData) (err error) {
			session := fsrpc.ConnSessionFromCtx(ctx)
			defer func() {
				if err != nil {
					log.Println("[AuthHandler] ServerCheck failed:", err, "remote=", session.RemoteAddr.String())
				}
			}()
			user := config.Users.Find(ar.GetUserName())
			if user == nil {
				return fmt.Errorf("user %q not found", ar.GetUserName())
			}
			return user.Check(ar)
		},
	}
}
