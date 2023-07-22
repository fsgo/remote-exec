// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/16

package packing

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/fsgo/fsgo/fsrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type User struct {
	Name     string
	PswMd5   string
	Psw      string
	Disabled bool
}

func (u *User) Check(ar *fsrpc.AuthData) error {
	now := time.Now()
	tm := ar.GetTimespan().AsTime()
	dl := now.Sub(tm)
	if dl > 5*time.Second || dl < -5*time.Second {
		return fmt.Errorf("invalid time %v", dl)
	}
	tokenWant := u.token(tm)
	if tokenWant != ar.GetToken() {
		return fmt.Errorf("invalid token %q", ar.GetToken())
	}
	return nil
}

func (u *User) getPswMd5() string {
	if u.PswMd5 != "" {
		return u.PswMd5
	}
	m5 := md5.New()
	m5.Write([]byte(u.Psw))
	return hex.EncodeToString(m5.Sum(nil))
}

func (u *User) ToAuthData() *fsrpc.AuthData {
	now := time.Now()
	return &fsrpc.AuthData{
		UserName: u.Name,
		Timespan: timestamppb.New(now),
		Token:    u.token(now),
	}
}

func (u *User) token(tm time.Time) string {
	m5 := md5.New()
	_, _ = m5.Write([]byte(u.Name))
	m5.Write([]byte(tm.UTC().Format(time.DateTime)))
	m5.Write([]byte(u.getPswMd5()))
	return hex.EncodeToString(m5.Sum(nil))
}

type Users []*User

func (us Users) Find(name string) *User {
	for _, u := range us {
		if !u.Disabled && u.Name == name {
			return u
		}
	}
	return nil
}
