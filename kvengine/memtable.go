package kvengine

//这里对于memtable暂时先考虑使用BST实现
//后面再考虑升级为RB tree

type Memtable interface {
	Set(kv Value) Status
	Search(key string) (*Value, Status)
	Delete(key string) Status
	GetAll() []Value
	GetCount() int
	Swap() Memtable
}
