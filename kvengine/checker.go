package kvengine

import "time"

//checker线程
//主要检查这几件事：
//1.内存表中是否满了需要向sstable迁移
//2.level中是否需要合并

func (DB *Database) Checker() {
	con := GetConfig()
	for {
		time.Sleep(time.Duration(con.CheckInterval) * time.Second)
		DB.MemCheck()
		DB.Tabletree.CheckCompaction()
	}
}
func (DB *Database) MemCheck() {
	con := GetConfig()
	if DB.Memtable.Count > con.Threshold {
		newbst := DB.Memtable.Swap()
		DB.Tabletree.CreateTable(newbst.GetAll())
		DB.Wal.Reset()
	}
}
