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
	for level := range tt.Levels {
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
func (tt *tabletree) Compaction(level int) {
	tablecache := make([]byte, levelMaxSize[level])
	currnode := tt.Levels[level]

	//用于合并的bst
	membst := NewBST()
	//把这一层的所有的value转移到bst中
	tt.Mu.Lock()
	for currnode != nil {
		table := currnode.Sstable

		if int64(len(tablecache)) < table.Tablemeta.Datalen {
			tablecache = make([]byte, table.Tablemeta.Datalen)
		}

		newslice := tablecache[0:table.Tablemeta.Datalen]

		_, err := table.F.Seek(0, 0)
		if err != nil {
			panic(err)
		}

		_, err = table.F.Read(newslice)
		if err != nil {
			panic(err)
		}

		for k, position := range table.Parseindexmap {
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
		currnode = currnode.Next
	}
	tt.Mu.Unlock()
	values := membst.GetAll()
	newlevel := level + 1
	if newlevel > 10 {
		newlevel = 10
	}
	//把这些所有的value合并到一个sstable中
	//并且放在下一层
	tt.CreateSstableAtLevel(values, newlevel)
	//清除当前层
	oldlevel := tt.Levels[level]
	if level < 10 {
		tt.Levels[level] = nil
		tt.ClearLevel(oldlevel)
	}
}

// 清除这一level
func (tt *tabletree) ClearLevel(node *tablenode) {
	tt.Mu.Lock()
	defer tt.Mu.Unlock()
	for node != nil {
		err := node.Sstable.F.Close()
		if err != nil {
			panic(err)
		}
		err = os.Remove(node.Sstable.Filepath)
		if err != nil {
			panic(err)
		}
		node.Sstable.F = nil
		node.Sstable = nil
		node = node.Next
	}
}
