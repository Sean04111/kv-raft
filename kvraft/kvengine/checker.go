package kvengine

import "time"

//checker线程
//主要检查这几件事：
//1.内存表中是否满了需要向sstable迁移
//2.level中是否需要合并

func (DB *Database)Checker() {
	con := GetConfig()
	ticker := time.Tick(time.Duration(con.CheckInterval) * time.Millisecond)
	for _ = range ticker {
		DB.MemCheck()
		DB.tabletree.CheckCompaction()
	}
}
func (DB *Database)MemCheck() {
	con := GetConfig()
	if DB.memtable.count > con.Threshold {
		newbst := DB.memtable.Swap()
		DB.tabletree.CreateTable(newbst.GetAll())
		DB.wal.Reset()
	}
}
