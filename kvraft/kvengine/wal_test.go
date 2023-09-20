package kvengine

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
)

func TestWALBasic(t *testing.T) {
	wal := new(wal)
	wal.Init("./wallogs/testwal.log")
	wal.Write(Value{"test", []byte("this is a test"), false})
	fmt.Println(wal.Init(wal.path).GetAll())
}
func TestWALMany(t *testing.T) {
	wal := new(wal)
	wal.Init("./wallogs/testwal.log")
	for i := 0; i < 10; i++ {
		wal.Write(Value{strconv.Itoa(i), []byte("this is a test"), false})
	}
	wal.Write(Value{"5",[]byte(""),true})
	fmt.Println(wal.Load().GetAll())
}
func TestA(t *testing.T) {
	s:="sstable/2.5.db"
	fmt.Println(filepath.Base(s))
}