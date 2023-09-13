package raft

import (
	"context"
	"testing"
	"time"
)

func TestRaft_DisplayLogBasic(t *testing.T) {
	peer := Raft{}
	peer.commitIndex = 3
	peer.lastApplied = 2
	peer.heartbeat = 50 * time.Millisecond
	peer.log.Entries = []Entry{{0, 0, -1}, {1, 1, -1}, {1, 2, "add"}, {1, 3, "del"}, {2, 4, "min"}, {2, 5, "squ"}}

	fctx, cancel := context.WithCancel(context.Background())
	sctx, _ := context.WithCancel(fctx)
	go peer.DisplayLog(sctx)

	time.Sleep(1 * time.Second)
	cancel()
}
func TestRaft_DisplayLogDynamicIndexDisplay(t *testing.T) {
	peer := Raft{}
	peer.commitIndex = 0
	peer.heartbeat = 50 * time.Millisecond
	peer.log.Entries = []Entry{{0, 0, -1}, {1, 1, -1}, {1, 2, "add"}, {1, 3, "del"}, {2, 4, "min"}, {2, 5, "squ"}}

	fctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	sctx1, _ := context.WithCancel(fctx)
	go peer.DisplayLog(sctx1)
	for {
		select {
		case <-fctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
			if peer.commitIndex < peer.LastIndex() {
				peer.commitIndex++
			}
		}
	}
}
func TestRaft_DisplayLogDynamicEntryDisplay(t *testing.T) {
	peer := Raft{}
	peer.commitIndex = 0
	peer.heartbeat = 50 * time.Millisecond
	peer.log.Entries = []Entry{{0, 0, -1}, {1, 1, -1}, {1, 2, "add"}, {1, 3, "del"}, {2, 4, "min"}, {2, 5, "squ"}}

	fctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	sctx1, _ := context.WithCancel(fctx)
	go peer.DisplayLog(sctx1)
	for {
		select {
		case <-fctx.Done():
			return
		default:
			time.Sleep(100 * time.Millisecond)
			peer.log.Entries = append(peer.log.Entries, Entry{1, peer.LastIndex() + 1, "adder"})
			peer.commitIndex++
		}
	}
}
func TestRaft_DisplayLogLargeLogDisplay(t *testing.T) {
	peer := Raft{}
	peer.commitIndex = 0
	peer.heartbeat = 50 * time.Millisecond
	peer.log.Entries = []Entry{{0, 0, -1}, {1, 1, -1}, {1, 2, "add"}, {1, 3, "del"}, {2, 4, "min"}, {2, 5, "squ"}}

	fctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	sctx1, _ := context.WithCancel(fctx)
	go peer.DisplayLog(sctx1)
	for {
		select {
		case <-fctx.Done():
			return
		default:
			time.Sleep(10 * time.Millisecond)
			peer.log.Entries = append(peer.log.Entries, Entry{1, peer.LastIndex() + 1, "adder"})
			peer.commitIndex++
		}
	}
}
