package kvengine

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestSkipListBasic(t *testing.T) {
	s := NewSkipList()
	s.Set(Value{Value: []byte("a"), Key: "a"})
	s.Set(Value{Value: []byte("b"), Key: "b"})
	fmt.Println(s.Search("a"))
	s.Delete("a")
	fmt.Println(s.Search("a"))
	fmt.Println(s.GetAll())
}
func BenchmarkNewSkipList_Set(b *testing.B) {
	s := NewSkipList()
	for i := 0; i < b.N; i++ {
		s.Set(Value{Value: []byte(strconv.Itoa(i)), Key: strconv.Itoa(i)})
	}
}
func BenchmarkSkipList_SearchAnsSet(b *testing.B) {
	s := NewSkipList()
	for i := 0; i < 100; i++ {
		s.Set(Value{Value: []byte(strconv.Itoa(i)), Key: strconv.Itoa(i)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Search(strconv.Itoa(i))
	}
}
func TestNewSkipList_Parral(t *testing.T) {
	s := NewSkipList()
	runtime.GOMAXPROCS(4)
	go func() {
		for i := 0; i < 100; i++ {
			s.Set(Value{Value: []byte(strconv.Itoa(i)), Key: strconv.Itoa(i)})
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			s.Delete(strconv.Itoa(i))
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			s.Search(strconv.Itoa(i))
		}
	}()
	time.Sleep(2 * time.Second)
}
