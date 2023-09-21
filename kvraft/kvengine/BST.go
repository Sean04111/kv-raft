package kvengine

import "sync"

type BSTNode struct {
	Val   Value
	Left  *BSTNode
	Right *BSTNode
}
type BST struct {
	root  *BSTNode
	count int
	mu    sync.RWMutex
}

// 这里要注意返回指针
// 不然会出现copy lock
func NewBST() *BST {
	bst := BST{}
	bst.root = nil
	bst.mu = sync.RWMutex{}
	return &bst
}

// 插入一个值
// 如果这个值本来已经被删除了,把deleted设置为false，并更新value
// 如果没有被删除,直接更新
// 这里的插入只允许插入到叶子节点
func (bst *BST) Set(kv Value) Status {
	bst.mu.Lock()
	defer bst.mu.Unlock()
	if bst == nil {
		return nilBST
	}
	newnode := &BSTNode{Val: kv, Left: nil, Right: nil}
	currnode := bst.root
	if currnode == nil {
		bst.root = newnode
		bst.count++
		return SetSuccess
	}
	//先找
	for currnode != nil {
		if currnode.Val.Key == kv.Key {
			currnode.Val.Value = kv.Value
			currnode.Val.Deleted = false
			return SetSuccess
		}
		if currnode.Val.Key < kv.Key {
			currnode = currnode.Right
		} else {
			currnode = currnode.Left
		}
	}
	//没找到
	nowNode := bst.root
	for nowNode != nil {
		//如果应在右边
		if nowNode.Val.Key < kv.Key {
			if nowNode.Right == nil {
				nowNode.Right = newnode
				bst.count++
				return SetSuccess
			} else {
				nowNode = nowNode.Right
			}
		} else {
			//应在左边
			if nowNode.Left == nil {
				nowNode.Left = newnode
				bst.count++
				return SetSuccess
			} else {
				nowNode = nowNode.Left
			}
		}
	}
	return SetFailed
}

// 搜索一个key
// 这里不能用递归,如果树的深度很深的话
// 压栈受不了
func (bst *BST) Search(key string) (*Value, Status) {
	bst.mu.RLock()
	defer bst.mu.RUnlock()
	if bst == nil {
		return nil, SearchNone
	}
	currnode := bst.root
	for currnode != nil {
		if currnode.Val.Key == key {
			if currnode.Val.Deleted {
				return nil, SearchDeleted
			} else {
				return &currnode.Val, SearchSuccess
			}
		}
		if currnode.Val.Key < key {
			currnode = currnode.Right
		} else {
			currnode = currnode.Left
		}
	}
	return nil, SearchNone
}

// 这里的删除只是改变字段，不是真的删除
func (bst *BST) Delete(key string) Status {
	bst.mu.Lock()
	defer bst.mu.Unlock()
	if bst == nil {
		return nilBST
	}
	find := bst.root
	for find != nil {
		if find.Val.Key == key {
			find.Val.Deleted = true
			bst.count--
			return DeleteSuccess
		}
		if find.Val.Key < key {
			find = find.Right
		} else {
			find = find.Left
		}
	}
	return DeleteNotFound
}

// 获取所有的值
// 不能压栈
// 中序遍历
func (bst *BST) GetAll() []Value {
	bst.mu.RLock()
	defer bst.mu.RUnlock()
	if bst == nil {
		return nil
	}
	cur, s := bst.root, []*BSTNode{}
	res := []Value{}
	for cur != nil || len(s) > 0 {
		for cur != nil {
			s = append(s, cur)
			cur = cur.Left
		}
		cur = s[len(s)-1]
		s = s[:len(s)-1]
		res = append(res, cur.Val)
		cur = cur.Right
	}
	return res
}

//内存表交换内存
func (bst *BST) Swap() *BST {
	bst.mu.Lock()
	defer bst.mu.Unlock()

	newbst := NewBST()
	newbst.root = bst.root
	bst.root = nil
	bst.count = 0
	return newbst
}
