// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package client

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fsgo/fsgo/fsrpc"
	"google.golang.org/protobuf/proto"

	"github.com/fsgo/remote-exec/internal/packing"
)

func Run() {
	parseFlag()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("no command to exec")
	}

	conn, err := config.Dial(time.Minute)
	checkErr(err)
	defer conn.Close()

	client := fsrpc.NewClient(conn)
	defer client.Close()

	rw := client.OpenStream()

	sendAuth(rw)
	sendExec(rw, args)
}

func sendExec(rw fsrpc.RequestWriter, args []string) {
	req := fsrpc.NewRequest("exec")
	req.LogID = strconv.FormatInt(time.Now().UnixNano(), 10)
	if debug {
		log.Println("send exec", args, "logid=", req.LogID)
	}

	py := &packing.Command{
		Name: args[0],
		Args: args[1:],
	}
	rr, err := fsrpc.WriteRequestProto(context.Background(), rw, req, py)
	checkErr(err)

	resp, payloads, err := rr.Response()
	checkErr(err)
	if code := resp.GetCode(); code != 0 {
		log.Fatalln("invalid response code", code, resp.GetMessage())
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
			_, _ = os.Stderr.Write([]byte("\n"))
			os.Exit(int(rt.ExitCode))
		}
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln("has error:", err)
	}
}
