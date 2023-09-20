package kvengine


//定义数据库实例
var database *Database
type Database struct{
	memtable *BST
	tabletree *tabletree
	wal *wal
}