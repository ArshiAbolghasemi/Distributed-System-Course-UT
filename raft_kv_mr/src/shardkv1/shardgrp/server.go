package shardgrp

import (
	"sync/atomic"

	"github.com/ArshiAbolghasemi/disgo/kvraft1/rsm"
	"github.com/ArshiAbolghasemi/disgo/kvsrv1/rpc"
	"github.com/ArshiAbolghasemi/disgo/labgob"
	"github.com/ArshiAbolghasemi/disgo/labrpc"
	"github.com/ArshiAbolghasemi/disgo/shardkv1/shardgrp/shardrpc"
	tester "github.com/ArshiAbolghasemi/disgo/tester1"
)

type KVServer struct {
	gid    tester.Tgid
	me     int
	dead   int32 // set by Kill()
	rsm    *rsm.RSM
	frozen bool // for testing purposes

}

func (kv *KVServer) DoOp(req any) any {
	// Your code here
	return nil
}

func (kv *KVServer) Snapshot() []byte {
	// Your code here
	return nil
}

func (kv *KVServer) Restore(data []byte) {
	// Your code here
}

func (kv *KVServer) Get(args *shardrpc.GetArgs, reply *rpc.GetReply) {
	// Your code here
}

func (kv *KVServer) Put(args *shardrpc.PutArgs, reply *rpc.PutReply) {
	// Your code here
}

// Freeze the specified shard (i.e., reject future Get/Puts for this
// shard) and return the key/values stored in that shard.
func (kv *KVServer) Freeze(args *shardrpc.FreezeArgs, reply *shardrpc.FreezeReply) {
	// Your code here
}

// Install the supplied state for the specified shard.
func (kv *KVServer) InstallShard(args *shardrpc.InstallShardArgs, reply *shardrpc.InstallShardReply) {
	// Your code here
}

// Delete the specified shard.
func (kv *KVServer) Delete(args *shardrpc.DeleteShardArgs, reply *shardrpc.DeleteShardReply) {
	// Your code here
}

// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	// Your code here, if desired.
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

// StartShardServerGrp starts a server for shardgrp `gid`.
//
// StartShardServerGrp() and MakeRSM() must return quickly, so they should
// start goroutines for any long-running work.
func StartServerShardGrp(servers []*labrpc.ClientEnd, gid tester.Tgid, me int, persister *tester.Persister, maxraftstate int) []tester.IService {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(shardrpc.PutArgs{})
	labgob.Register(shardrpc.GetArgs{})
	labgob.Register(shardrpc.FreezeArgs{})
	labgob.Register(shardrpc.InstallShardArgs{})
	labgob.Register(shardrpc.DeleteShardArgs{})
	labgob.Register(rsm.Op{})

	kv := &KVServer{gid: gid, me: me}
	kv.rsm = rsm.MakeRSM(servers, me, persister, maxraftstate, kv)

	// Your code here
	return []tester.IService{kv, kv.rsm.Raft()}
}
