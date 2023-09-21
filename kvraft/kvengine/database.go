package kvengine

import (
	"encoding/json"
	"fmt"
	"kv-raft/kvraft"
	"os"
)

//定义数据库实例
//全局的DB对象
//这里作为实现statemachine的接口

type Database struct {
	memtable  *BST
	tabletree *tabletree
	wal       *wal
}

func Start(con Config) *Database {
	db := &Database{
		memtable:  NewBST(),
		tabletree: &tabletree{},
		wal:       &wal{},
	}
	ConfigInit(con)
	db.InitDB(con.DataDir)
	db.MemCheck()
	db.tabletree.CheckCompaction()
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
	mem := db.wal.Init(dir)
	db.memtable = mem
	db.tabletree.Init(dir)
}
func (db *Database) Get(key string) (string, kvraft.Err) {
	//先在内存表里面找
	val, stat := db.memtable.Search(key)
	if stat == SearchSuccess {
		return string(val.Value), kvraft.OK
	}
	//如果没找到，在sstable里面找
	if db.tabletree != nil {
		v, s := db.tabletree.Search(key)
		if s == SearchSuccess {
			return string(v.Value), kvraft.OK
		}
	}
	return kvraft.EmptyString, kvraft.ErrNoKey
}
func (db *Database) Put(key, value string) kvraft.Err {
	data, _ := json.Marshal(value)
	db.memtable.Set(Value{Key: key, Value: data, Deleted: false})

	db.wal.Write(Value{Key: key, Value: data, Deleted: false})
	return kvraft.OK
}
func (db *Database) Append(key, value string) kvraft.Err {
	val, _ := db.Get(key)
	newval := fmt.Sprintf("%s%s",val,value)
	fmt.Println("get",val)
	fmt.Println("newval", newval)
	db.Put(key, newval)
	return kvraft.OK
}
