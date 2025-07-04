package shardgrp

import (
	"github.com/ArshiAbolghasemi/disgo/kvsrv1/rpc"
	"github.com/ArshiAbolghasemi/disgo/shardkv1/shardcfg"
	tester "github.com/ArshiAbolghasemi/disgo/tester1"
)

type Clerk struct {
	clnt    *tester.Clnt
	servers []string
	leader  int // last successful leader (index into servers[])
}

func MakeClerk(clnt *tester.Clnt, servers []string) *Clerk {
	ck := &Clerk{clnt: clnt, servers: servers}
	return ck
}

func (ck *Clerk) Get(key string, n shardcfg.Tnum) (string, rpc.Tversion, rpc.Err) {
	// Your code here
	return "", 0, ""
}

func (ck *Clerk) Put(key string, value string, version rpc.Tversion, n shardcfg.Tnum) (bool, rpc.Err) {
	// Your code here
	return false, ""
}

func (ck *Clerk) Freeze(s shardcfg.Tshid, num shardcfg.Tnum) ([]byte, rpc.Err) {
	return nil, ""
}

func (ck *Clerk) InstallShard(s shardcfg.Tshid, state []byte, num shardcfg.Tnum) rpc.Err {
	return ""
}

func (ck *Clerk) Delete(s shardcfg.Tshid, num shardcfg.Tnum) rpc.Err {
	return ""
}
