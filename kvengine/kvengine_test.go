package kvengine

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestBasi(t *testing.T) {
	db := Start(Config{
		DataDir:       "./dbfiles",
		Level0Size:    100,
		PartSize:      4,
		Threshold:     30,
		CheckInterval: 3,
	})
	for i := 0; i < 100; i++ {
		db.Put(strconv.Itoa(i), fmt.Sprint(i))
	}
	fmt.Println(db.Get("19"))

}

// 测试连续插入100万条数据
// 中间sleep是为了checker
func TestMuch(t *testing.T) {
	db := Start(Config{
		DataDir:       "./dbfiles",
		Level0Size:    5,
		PartSize:      2,
		Threshold:     300,
		CheckInterval: 3,
	})
	for i := 0; i < 200000; i++ {
		db.Put(strconv.Itoa(i), strconv.Itoa(i))
	}
	time.Sleep(1 * time.Second)
	for i := 200000; i < 400000; i++ {
		db.Put(strconv.Itoa(i), strconv.Itoa(i))
	}
	time.Sleep(1 * time.Second)
	for i := 400000; i < 600000; i++ {
		db.Put(strconv.Itoa(i), strconv.Itoa(i))
	}
	time.Sleep(1 * time.Second)
	for i := 600000; i < 800000; i++ {
		db.Put(strconv.Itoa(i), strconv.Itoa(i))
	}
	time.Sleep(1 * time.Second)
	for i := 800000; i < 1000000; i++ {
		db.Put(strconv.Itoa(i), strconv.Itoa(i))
	}
}

// 测试热启动
func TestHotStart(t *testing.T) {
	db := Start(Config{
		DataDir:       "./dbfiles",
		Level0Size:    100,
		PartSize:      4,
		Threshold:     3000,
		CheckInterval: 3,
	})
	fmt.Println(db.Get("56462"))
}

// 测试序列化编码
func TestMarshal(t *testing.T) {
	db := Start(Config{
		DataDir:       "./dbfiles",
		Level0Size:    5,
		PartSize:      2,
		Threshold:     300,
		CheckInterval: 3,
	})
	db.Put("1", "testo")
	fmt.Println(db.Get("1"))
}

func TestAppend(t *testing.T) {
	db := Start(Config{
		DataDir:       "./dbfiles",
		Level0Size:    5,
		PartSize:      2,
		Threshold:     300,
		CheckInterval: 3,
	})
	db.Append("append","this is append")
	fmt.Println(db.Get("append"))
}
