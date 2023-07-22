// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/fsgo/fsgo/fsrpc"

	"github.com/fsgo/remote-exec/internal/packing"
)

var _ fsrpc.Handler = (*Handler)(nil)

type Handler struct {
	cfg *Config
}

func (h *Handler) Handle(ctx context.Context, rr fsrpc.RequestReader, rw fsrpc.ResponseWriter) (err error) {
	req, rc, err := fsrpc.ReadRequestProto(ctx, rr, &packing.Command{})
	{
		session := fsrpc.ConnSessionFromCtx(ctx)
		logMsg := fmt.Sprintf("remote=%s, logid=%s, cmd=%s %q",
			session.RemoteAddr.String(),
			req.GetLogID(),
			rc.GetName(),
			rc.GetArgs(),
		)
		start := time.Now()
		log.Println("handler start", logMsg)
		defer func() {
			log.Println("handler finish,", logMsg, ",cost=", time.Since(start), ",err=", err)
		}()
	}

	if err != nil {
		resp := fsrpc.NewResponse(req.GetID(), fsrpc.ErrCode_Internal, err.Error())
		_ = fsrpc.WriteResponseProto(ctx, rw, resp)
		return err
	}

	resp := fsrpc.NewResponseSuccess(req.GetID())
	pc := fsrpc.PayloadChan[*packing.CommandResult]{
		RID:          req.GetID(),
		EncodingType: fsrpc.EncodingType_Protobuf,
	}

	cmdCtx := ctx // 执行 Command 专用的 ctx
	cmdCtx, cmdCancel := context.WithCancelCause(cmdCtx)
	go func() {
		err2 := rw.Write(ctx, resp, pc.Chan())
		if err2 != nil {
			cmdCancel(err2)
			log.Println("write response chan error:", err2)
		}
	}()

	if !h.cfg.CmdAllow(rc.GetName()) {
		rt := &packing.CommandResult{
			Stderr: []byte(fmt.Sprintf("command %s now allow", rc.GetName())),
		}
		pc.Write(ctx, rt, false)
		return nil
	}

	timeout := rc.GetTimeoutMS()

	if timeout > 0 {
		var cancel context.CancelFunc
		cmdCtx, cancel = context.WithTimeout(cmdCtx, time.Duration(timeout)*time.Millisecond)
		defer cancel()
	}

	cmd := exec.CommandContext(cmdCtx, rc.GetName(), rc.GetArgs()...)
	cmd.Stdout = &xWriter{
		onWrite: func(p []byte) error {
			rt := &packing.CommandResult{
				Stdout: p,
			}
			err1 := pc.Write(ctx, rt, true)
			if err1 != nil {
				log.Println("Stdout write:", err1)
				cmdCancel(err1)
			}
			return err1
		},
	}
	cmd.Stderr = &xWriter{
		onWrite: func(p []byte) error {
			rt := &packing.CommandResult{
				Stderr: p,
			}
			err1 := pc.Write(ctx, rt, true)
			if err1 != nil {
				log.Println("Stderr write:", err1)
				cmdCancel(err1)
			}
			return err1
		},
	}
	err = cmd.Run()
	rt := &packing.CommandResult{
		Finished: true,
		ExitCode: int32(cmd.ProcessState.ExitCode()),
	}
	if err != nil {
		rt.Stderr = []byte(err.Error())
	}
	_ = pc.Write(ctx, rt, false)
	return err
}

var _ io.Writer = (*xWriter)(nil)

type xWriter struct {
	onWrite func(p []byte) error
}

func (xw *xWriter) Write(p []byte) (n int, err error) {
	if err = xw.onWrite(p); err != nil {
		return 0, err
	}
	return len(p), nil
}
