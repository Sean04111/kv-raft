package kvengine

import "time"

//checker线程
//主要检查这几件事：
//1.内存表中是否满了需要向sstable迁移
//2.level中是否需要合并

func Checker(){
	con:=GetConfig()
	ticker:=time.Tick(time.Duration(con.CheckInterval)*time.Second)
	for _=range ticker{
		MemCheck()
		database.tabletree.CheckCompaction()
	}
}
func MemCheck(){
	con:=GetConfig()
	if database.memtable.count>con.Threshold{
		newbst := database.memtable.Swap()
		database.tabletree.CreateTable(newbst.GetAll())
		database.wal.Reset()
	}
}