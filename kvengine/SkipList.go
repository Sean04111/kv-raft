package kvengine

import (
	"math/rand"
	"sync"
)

// SkipList
// 对于memtable的跳表实现
type SkipList struct {
	head     *SkipListNode
	objcache sync.Pool
	count    int
	rwmu     sync.RWMutex
}
type SkipListNode struct {
	nexts []*SkipListNode
	val   Value
}

func NewSkipList() *SkipList {
	sl := &SkipList{}
	sl.head = &SkipListNode{}
	sl.count = 0
	sl.objcache = sync.Pool{New: func() any {
		return &SkipListNode{}
	}}
	return sl
}
func (sl *SkipList) Search(key string) (*Value, Status) {
	sl.rwmu.RLock()
	defer sl.rwmu.RUnlock()
	node := sl.head
	for level := len(sl.head.nexts) - 1; level >= 0; level-- {
		//当前层数开始一直找
		for node.nexts[level] != nil && node.nexts[level].val.Key < key {
			node = node.nexts[level]
		}
		if node.nexts[level] != nil && node.nexts[level].val.Key == key {
			return &node.nexts[level].val, SearchSuccess
		}
	}
	return nil, SearchNone
}

func (sl *SkipList) roll() int {
	var level int
	for rand.Intn(2) > 0 {
		level++
	}
	return level
}
func (sl *SkipList) Set(kv Value) Status {
	//先查看是否是更新操作
	//先不上锁
	got, res := sl.Search(kv.Key)
	if res == SearchSuccess {
		sl.rwmu.Lock()
		got.Key = kv.Key
		got.Value = kv.Value
		got.Deleted = false
		sl.rwmu.Unlock()
		return SetSuccess
	} else {
		//是插入操作
		sl.rwmu.Lock()
		//先随机出这个node的层数
		level := sl.roll()
		newnode := sl.objcache.Get().(*SkipListNode)
		newnode.nexts = make([]*SkipListNode, level+1)
		newnode.val = kv

		//如果层数太大要给head扩容
		for level > len(sl.head.nexts)-1 {
			sl.head.nexts = append(sl.head.nexts, nil)
		}

		//开始set
		move := sl.head
		for level := level; level >= 0; level-- {
			for move.nexts[level] != nil && move.nexts[level].val.Key < newnode.val.Key {
				move = move.nexts[level]
			}
			//插入新节点
			newnode.nexts[level] = move.nexts[level]
			move.nexts[level] = newnode
		}
		sl.count++
		sl.rwmu.Unlock()
		return SetSuccess
	}
}
func (sl *SkipList) Delete(key string) Status {
	_, ok := sl.Search(key)
	if ok != SearchSuccess {
		return DeleteNotFound
	}
	sl.rwmu.Lock()
	move := sl.head
	for l := len(sl.head.nexts) - 1; l >= 0; l-- {
		for move.nexts[l] != nil && move.nexts[l].val.Key < key {
			move = move.nexts[l]
		}
		if move.nexts[l] != nil && move.nexts[l].val.Key == key {
			dn := move.nexts[l]
			move.nexts[l] = move.nexts[l].nexts[l]
			sl.objcache.Put(dn)
		}
	}

	//更新层的高度
	emptylevel := 0
	for l := len(sl.head.nexts) - 1; l > 0 && sl.head.nexts[l] == nil; l-- {
		emptylevel++
	}
	sl.head.nexts = sl.head.nexts[:len(sl.head.nexts)-emptylevel]
	sl.count--
	sl.rwmu.Unlock()
	return DeleteSuccess
}
func (sl *SkipList) GetAll() []Value {
	sl.rwmu.RLock()
	defer sl.rwmu.RUnlock()
	var ans []Value
	p := sl.head.nexts[0]
	for p != nil {
		ans = append(ans, p.val)
		p = p.nexts[0]
	}
	return ans
}
func (sl *SkipList) GetCount() int {
	sl.rwmu.RLock()
	ans := sl.count
	sl.rwmu.RUnlock()
	return ans
}

// 把sl中的数据换出来
func (sl *SkipList) Swap() Memtable {
	sl.rwmu.Lock()
	defer sl.rwmu.Unlock()

	newsl := NewSkipList()
	newsl.head = sl.head
	newsl.count = sl.count

	sl.head = &SkipListNode{}
	sl.count = 0
	return newsl
}
