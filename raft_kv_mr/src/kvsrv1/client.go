package kvsrv

import (
	"time"

	"github.com/ArshiAbolghasemi/disgo/kvsrv1/rpc"
	kvtest "github.com/ArshiAbolghasemi/disgo/kvtest1"
	tester "github.com/ArshiAbolghasemi/disgo/tester1"
)

type Clerk struct {
	clnt   *tester.Clnt
	server string
}

func MakeClerk(clnt *tester.Clnt, server string) kvtest.IKVClerk {
	ck := &Clerk{clnt: clnt, server: server}
	// You may add code here.
	return ck
}

func (ck *Clerk) wait() {
	time.Sleep(100 * time.Millisecond)
}

// Get fetches the current value and version for a key.  It returns
// ErrNoKey if the key does not exist. It keeps trying forever in the
// face of all other errors.
//
// You can send an RPC with code like this:
// ok := ck.clnt.Call(ck.server, "KVServer.Get", &args, &reply)
//
// The types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. Additionally, reply must be passed as a pointer.
func (ck *Clerk) Get(key string) (string, rpc.Tversion, rpc.Err) {
	// You will have to modify this function.
	args := rpc.GetArgs{
		Key: key,
	}
	for {
		reply := rpc.GetReply{}
		ok := ck.clnt.Call(ck.server, "KVServer.Get", &args, &reply)
		if ok {
			return reply.Value, reply.Version, reply.Err
		}
		ck.wait()
	}
}

// Put updates key with value only if the version in the
// request matches the version of the key at the server.  If the
// versions numbers don't match, the server should return
// ErrVersion.  If Put receives an ErrVersion on its first RPC, Put
// should return ErrVersion, since the Put was definitely not
// performed at the server. If the server returns ErrVersion on a
// resend RPC, then Put must return ErrMaybe to the application, since
// its earlier RPC might have been processed by the server successfully
// but the response was lost, and the Clerk doesn't know if
// the Put was performed or not.
//
// You can send an RPC with code like this:
// ok := ck.clnt.Call(ck.server, "KVServer.Put", &args, &reply)
//
// The types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. Additionally, reply must be passed as a pointer.
func (ck *Clerk) Put(key, value string, version rpc.Tversion) rpc.Err {
	// You will have to modify this function.
	args := rpc.PutArgs{
		Key:     key,
		Value:   value,
		Version: version,
	}
	maybe := false
	for {
		reply := rpc.PutReply{}
		ok := ck.clnt.Call(ck.server, "KVServer.Put", &args, &reply)
		if !ok {
			// we might have successfully put this key into the server but the response was lost
			maybe = true
			ck.wait()
		} else {
			// if we got `ErrVersion` with maybe, then this is not the first time we send it.
			// we need to return `ErrMaybe` since we don't know the previous rpc was success or not
			if maybe && reply.Err == rpc.ErrVersion {
				return rpc.ErrMaybe
			}
			return reply.Err
		}

	}
}
