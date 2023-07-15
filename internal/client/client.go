// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package client

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"time"

	"github.com/fsgo/fsgo/fsrpc"
	"google.golang.org/protobuf/proto"

	"github.com/fsgo/remote-exec/internal/packing"
)

func Run(addr string) {
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("no command to exec")
	}
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	checkErr(err)

	defer conn.Close()
	client := fsrpc.NewClientConn(conn)

	rw := client.MustOpen(context.Background())
	req := fsrpc.NewRequest("exec")
	py := &packing.Command{
		Name: args[0],
		Args: args[1:],
	}
	rr, err := fsrpc.WriteRequestProto(context.Background(), rw, req, py)
	checkErr(err)

	resp, payloads, err := rr.Response()
	checkErr(err)
	if code := resp.GetCode(); code != 0 {
		log.Fatalln(resp.GetMessage())
	}
	for item := range payloads {
		bf, err1 := item.Bytes()
		checkErr(err1)
		rt := &packing.CommandResult{}
		err2 := proto.Unmarshal(bf, rt)
		checkErr(err2)
		if len(rt.Stdout) > 0 {
			_, _ = os.Stdout.Write(rt.Stdout)
		}
		if len(rt.Stderr) > 0 {
			_, _ = os.Stderr.Write(rt.Stderr)
		}
		if rt.Finished {
			os.Exit(int(rt.ExitCode))
		}
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
