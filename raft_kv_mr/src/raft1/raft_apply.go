package raft

import (
	"github.com/ArshiAbolghasemi/disgo/raftapi"
)

// looping and applying log to the state machine
func (rf *Raft) applier() {
	for {
		select {
		case <-rf.commitIndexChangedCh:
			rf.applyLog()
		case <-rf.snapshotChangedCh:
			rf.applySnapshot()
		case <-rf.closeApplyCh:
			close(rf.applyCh)
			return
		}
	}
}

func (rf *Raft) applyLog() {
	logs := make(raftLog, 0)

	rf.lock("applyLog")
	if rf.lastApplied < rf.commitIndex {
		logs = rf.log[rf.lastApplied+1 : rf.commitIndex+1]
		rf.lastApplied = rf.commitIndex
	}
	rf.unlock("applyLog")

	for _, log := range logs {
		if rf.killed() {
			return
		}

		switch log.LogEntryType {

		case clientLogEntry:
			// wait on blocking channel, to avoid sending command out of order
			msg := raftapi.ApplyMsg{
				CommandValid: true,
				Command:      log.Command,
				CommandIndex: log.CommandIndex,
			}
			rf.applyCh <- msg
			DPrintf(tApply, "S%d(%d,-), apply, log: %+v", rf.me, log.Term, log)

		case noOpLogEntry:
			DPrintf(tApply, "S%d(%d,-), skip, log: %+v", rf.me, log.Term, log)
		}
	}
}

func (rf *Raft) applySnapshot() {
	rf.lock("applySnapshot")
	snapshot := make([]byte, len(rf.snapshot))
	copy(snapshot, rf.snapshot)
	msg := raftapi.ApplyMsg{
		SnapshotValid: true,
		Snapshot:      snapshot,
		SnapshotTerm:  rf.log[0].Term,
		SnapshotIndex: rf.log[0].CommandIndex,
	}
	rf.unlock("applySnapshot")

	// wait on blocking channel, to avoid sending command out of order
	rf.applyCh <- msg
	DPrintf(tApply, "S%d(%d,-), apply, snapshot: %+v", rf.me, msg.SnapshotTerm, msg)
}
