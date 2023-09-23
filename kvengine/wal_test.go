package kvengine

import (
	"fmt"
	"strconv"
	"testing"
)

func TestWALBasic(t *testing.T) {
	wal := new(wal)
	wal.Init("./wallogs/testwal.log")
	wal.Write(Value{"test", []byte("this is a test"), false})
	fmt.Println(wal.Init(wal.Path).GetAll())
}
func TestWALMany(t *testing.T) {
	wal := new(wal)
	wal.Init("./wallogs/testwal.log")
	for i := 0; i < 10; i++ {
		wal.Write(Value{strconv.Itoa(i), []byte("this is a test"), false})
	}
	wal.Write(Value{"5", []byte(""), true})
	fmt.Println(wal.Load().GetAll())
}

func TestXxx(t *testing.T) {
	s := "abc"
	ans:=0
	for i := range s {
		m := s[i]
		ans+=int(m)
	}
	fmt.Println(ans)
}
