// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package server

import (
	"flag"
	"log"

	"github.com/fsgo/fsgo/fsrpc"
)

func Run() {
	flag.Parse()
	router := fsrpc.NewRouter()
	router.Register("exec", &Handler{})
	log.Println("server Listen at:", config.GetAddr())

	var err error
	defer func() {
		log.Println("server exit:", err)
	}()
	err = fsrpc.ListenAndServe(config.GetAddr(), router)
}
