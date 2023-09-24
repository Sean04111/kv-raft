package kvengine

import (
	"fmt"
	"kv-raft/common"
	"os"
)

//定义数据库实例
//全局的DB对象
//这里作为实现statemachine的接口

type Database struct {
	Memtable  Memtable
	Tabletree *tabletree
	Wal       *wal
}

func Start(con Config) *Database {
	db := &Database{
		Memtable:  NewBST(),
		Tabletree: &tabletree{},
		Wal:       &wal{},
	}
	ConfigInit(con)
	db.InitDB(con.DataDir)
	db.MemCheck()
	db.Tabletree.CheckCompaction()
	go db.Checker()
	return db
}
func (db *Database) InitDB(dir string) {
	_, err := os.Stat(dir)
	//目录不存在,之间返回空数据库
	if err != nil {
		err = os.Mkdir(dir, 0666)
		if err != nil {
			panic(err)
		}
	}
	mem := db.Wal.Init(dir)
	db.Memtable = mem
	db.Tabletree.Init(dir)
}
func (db *Database) Get(key string) (string, common.Err) {
	//先在内存表里面找
	val, stat := db.Memtable.Search(key)
	if stat == SearchSuccess {
		return string(val.Value), common.OK
	}
	//如果没找到，在sstable里面找
	if db.Tabletree != nil {
		v, s := db.Tabletree.Search(key)
		if s == SearchSuccess {
			return string(v.Value), common.OK
		}
	}
	return common.EmptyString, common.ErrNoKey
}
func (db *Database) Put(key, value string) common.Err {
	data := []byte(value)
	db.Memtable.Set(Value{Key: key, Value: data, Deleted: false})

	db.Wal.Write(Value{Key: key, Value: data, Deleted: false})
	return common.OK
}
func (db *Database) Append(key, value string) common.Err {
	val, _ := db.Get(key)
	newval := fmt.Sprintf("%s%s", val, value)
	db.Put(key, newval)
	return common.OK
}
