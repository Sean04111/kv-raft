package kvraft

import (
	"fmt"
	"strconv"
	"testing"
)

func TestBasic(t *testing.T) {
	var kvstate KVStateMachine
	kvstate = NewMKV()
	for i := 0; i < 10; i++ {
		kvstate.Put(strconv.Itoa(i), strconv.Itoa(i))
	}
	kvstate.Append(strconv.Itoa(6), "this is 6")
	for i := 0; i < 10; i++ {
		fmt.Println(kvstate.Get(strconv.Itoa(i)))
	}
}
