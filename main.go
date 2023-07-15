// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package main

import (
	"flag"

	"github.com/fsgo/remote-exec/internal/client"
)

var serverAddr = flag.String("addr", "127.0.0.1:8100", "server addr")

func main() {
	flag.Parse()
	client.Run(*serverAddr)
}
