package kvengine

import (
	"fmt"
	"strconv"
	"testing"
)

func TestBasic(t *testing.T) {
	testbst := NewBST()
	for i := 0; i < 10; i++ {
		testbst.Set(Value{Key: strconv.Itoa(i), Value: []byte(strconv.Itoa(i))})
	}
	fmt.Println(testbst.GetAll())
	fmt.Println(testbst.Search("1"))
	fmt.Println(testbst.Search("10"))
	fmt.Println(testbst.Delete("1"))
	testbst.Set(Value{"1", []byte("new1"), false})
}
func TestMany(t *testing.T) {
	testbst := NewBST()
	for i := 0; i < 100; i++ {
		testbst.Set(Value{Key: strconv.Itoa(i), Value: []byte(strconv.Itoa(i))})
	}
	for i := 0; i < 100; i++ {
		fmt.Println(testbst.Search(strconv.Itoa(i)))
	}
	for i := 0; i < 100; i++ {
		testbst.Delete(strconv.Itoa(i))
	}
}
func TestParall(t *testing.T) {
	testbst := NewBST()
	for i := 0; i < 10; i++ {
		go func() {
			testbst.Set(Value{Key: strconv.Itoa(i), Value: []byte(strconv.Itoa(i))})
		}()
	}
	for i := 0; i < 10; i++ {
		fmt.Println(testbst.Delete(strconv.Itoa(i)))
	}

}
