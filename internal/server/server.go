// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package server

import (
	"flag"
	"log"

	"github.com/fsgo/fsgo/fsrpc"
)

var config = &Config{}

func init() {
	config.RegisterFlag()
}

func Run() {
	flag.Parse()

	c1, err := parserCnf()
	checkErr(err)
	config.Merge(c1)

	router := fsrpc.NewRouter()

	auth := serverAuth()
	auth.RegisterTo(router)

	router.Register("exec", auth.WithInterceptor(&Handler{}))
	// router.Register("exec",&Handler{})

	defer func() {
		log.Println("server exit:", err)
	}()
	ser := &fsrpc.Server{
		Router: router,
	}
	err = config.Run(ser)
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
