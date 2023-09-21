package kvengine

import (
	"encoding/json"
	"fmt"
	"os"
)

//sstable之间的合并
//在程序启动后，启动一个线程来检查

// 检查这个table tree是否需要压缩
func (tt *tabletree) CheckCompaction() {
	for level := range tt.levels {
		levelsize := tt.GetLevelSize(level) / 1000 / 1000
		levelnum := tt.GetLevelNum(level)
		con := GetConfig()
		//判断当前level是否需要压缩
		if levelnum > con.PartSize || levelsize > int64(levelMaxSize[level]) {
			tt.Compaction(level)
		}
	}
}

// 压缩当前level文件到下一层
// 合并的思路是使用bst
// 数据合并有问题
//这个合并操作只能做一次,第二次就会打乱数据
func (tt *tabletree) Compaction(level int) {
	tablecache := make([]byte, levelMaxSize[level])
	currnode := tt.levels[level]

	//用于合并的bst
	membst := NewBST()
	//把这一层的所有的value转移到bst中
	tt.mu.Lock()
	for currnode != nil {
		table := currnode.sstable

		if int64(len(tablecache)) < table.tablemeta.datalen {
			tablecache = make([]byte, table.tablemeta.datalen)
		}

		newslice := tablecache[0:table.tablemeta.datalen]

		_, err := table.f.Seek(0, 0)
		if err != nil {
			panic(err)
		}

		_, err = table.f.Read(newslice)
		if err != nil {
			panic(err)
		}

		for k, position := range table.parseindexmap {
			if position.Deleted == false {
				var v Value
				er := json.Unmarshal(newslice[position.Start:position.Start+position.Len], &v)
				if er != nil {
					fmt.Println(er)
				}
				membst.Set(v)
			} else {
				membst.Delete(k)
			}
		}
		currnode = currnode.next
	}
	tt.mu.Unlock()
	values := membst.GetAll()
	newlevel := level + 1
	if newlevel > 10 {
		newlevel = 10
	}
	//把这些所有的value合并到一个sstable中
	//并且放在下一层
	tt.CreateSstableAtLevel(values, newlevel)
	//清除当前层
	oldlevel := tt.levels[level]
	if level < 10 {
		tt.levels[level] = nil
		tt.ClearLevel(oldlevel)
	}
}

// 清除这一level
//会报错：0.0db正在被占用
func (tt *tabletree) ClearLevel(node *tablenode) {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	for node != nil {
		err := node.sstable.f.Close()
		if err != nil {
			panic(err)
		}
		err = os.Remove(node.sstable.filepath)
		if err != nil {
			panic(err)
		}
		node.sstable.f = nil
		node.sstable = nil
		node = node.next
	}
}
