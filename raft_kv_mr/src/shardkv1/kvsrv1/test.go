package kvsrv

import (
	// "log"
	"testing"

	kvtest "github.com/ArshiAbolghasemi/disgo/kvtest1"
	tester "github.com/ArshiAbolghasemi/disgo/tester1"
)

type TestKV struct {
	*kvtest.Test
	t        *testing.T
	reliable bool
}

func MakeTestKV(t *testing.T, reliable bool) *TestKV {
	cfg := tester.MakeConfig(t, 1, reliable, StartKVServer)
	ts := &TestKV{
		t:        t,
		reliable: reliable,
	}
	ts.Test = kvtest.MakeTest(t, cfg, false, ts)
	return ts
}

func (ts *TestKV) MakeClerk() kvtest.IKVClerk {
	clnt := ts.Config.MakeClient()
	ck := MakeClerk(clnt, tester.ServerName(tester.GRP0, 0))
	return &kvtest.TestClerk{ck, clnt}
}

func (ts *TestKV) DeleteClerk(ck kvtest.IKVClerk) {
	tck := ck.(*kvtest.TestClerk)
	ts.DeleteClient(tck.Clnt)
}
