// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package server

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/fsgo/fsgo/fsrpc"

	"github.com/fsgo/remote-exec/internal/packing"
)

var _ fsrpc.Handler = (*Handler)(nil)

type Handler struct {
	cfg *Config
}

func (h *Handler) Handle(ctx context.Context, rr fsrpc.RequestReader, rw fsrpc.ResponseWriter) error {
	req, rc, err := fsrpc.ReadRequestProto(ctx, rr, &packing.Command{})
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

	go func() {
		_ = rw.Write(ctx, resp, pc.Chan())
	}()

	if !h.cfg.CmdAllow(rc.GetName()) {
		rt := &packing.CommandResult{
			Stderr: []byte(fmt.Sprintf("command %s now allow", rc.GetName())),
		}
		pc.Write(ctx, rt, false)
	}

	timeout := rc.GetTimeoutMS()

	cmdCtx := ctx // 执行 Command 专用的 ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		cmdCtx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
		defer cancel()
	}
	cmd := exec.CommandContext(cmdCtx, rc.GetName(), rc.GetArgs()...)
	cmd.Stdout = &xWriter{
		onWrite: func(p []byte) {
			rt := &packing.CommandResult{
				Stdout: p,
			}
			pc.Write(ctx, rt, false)
		},
	}
	cmd.Stderr = &xWriter{
		onWrite: func(p []byte) {
			rt := &packing.CommandResult{
				Stderr: p,
			}
			pc.Write(ctx, rt, false)
		},
	}
	err = cmd.Run()
	rt := &packing.CommandResult{
		Finished: true,
		ExitCode: int32(cmd.ProcessState.ExitCode()),
	}
	_ = pc.Write(ctx, rt, false)
	return err
}

var _ io.Writer = (*xWriter)(nil)

type xWriter struct {
	onWrite func(p []byte)
}

func (xw *xWriter) Write(p []byte) (n int, err error) {
	xw.onWrite(p)
	return len(p), nil
}
