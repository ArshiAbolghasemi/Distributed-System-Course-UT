package shardrpc

import (
	"github.com/ArshiAbolghasemi/disgo/kvsrv1/rpc"
	"github.com/ArshiAbolghasemi/disgo/shardkv1/shardcfg"
)

// Same as Put in kvsrv1/rpc, but with a configuration number
type PutArgs struct {
	Key     string
	Value   string
	Version rpc.Tversion
	Num     shardcfg.Tnum
}

// Same as Get in kvsrv1/rpc, but with a configuration number.
type GetArgs struct {
	Key string
	Num shardcfg.Tnum
}

type FreezeArgs struct {
	Shard shardcfg.Tshid
	Num   shardcfg.Tnum
}

type FreezeReply struct {
	State []byte
	Num   shardcfg.Tnum
	Err   rpc.Err
}

type InstallShardArgs struct {
	Shard shardcfg.Tshid
	State []byte
	Num   shardcfg.Tnum
}

type InstallShardReply struct {
	Err rpc.Err
}

type DeleteShardArgs struct {
	Shard shardcfg.Tshid
	Num   shardcfg.Tnum
}

type DeleteShardReply struct {
	Err rpc.Err
}
