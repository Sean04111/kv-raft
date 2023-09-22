package kvengine

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

// 把sstable分成三个区，数据区、索引区、元数据区
type metaInfo struct {
	//版本号？
	Version int64
	//数据区起点
	Datastart int64
	//数据区长度
	Datalen int64
	//索引区起点
	Indexstart int64
	//索引区长度
	Indexlen int64
}

// 储存value的block的位置信息
// 在sstable中map和key对应
type Position struct {
	Start   int64
	Len     int64
	Deleted bool
}

// sstable磁盘文件
type sstable struct {
	Mu sync.Mutex
	//文件句柄
	F        *os.File
	Filepath string

	//元数据
	Tablemeta metaInfo
	//索引
	Parseindexmap map[string]Position
	//排好序的key列表
	Sortedkeys []string
}

func (t *sstable) Init(path string) {
	t.Filepath = path
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	t.F = f
	//load元数据和索引
	//只需要meta和index就可以推出data的区域
	//顺序不能变
	t.LoadMeta()

	t.LoadIndex()
}

func (t *sstable) LoadIndex() {
	bytes := make([]byte, t.Tablemeta.Indexlen)
	t.F.Seek(t.Tablemeta.Indexstart, 0)
	_, err := t.F.Read(bytes)
	if err != nil {
		panic(err)
	}
	t.Parseindexmap = map[string]Position{}
	err = json.Unmarshal(bytes, &t.Parseindexmap)
	if err != nil {
		fmt.Println(err)
	}
	t.F.Seek(0, 0)

	keys := []string{}
	for k := range t.Parseindexmap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	t.Sortedkeys = keys

}
func (t *sstable) LoadMeta() {
	info, err := t.F.Stat()
	if err != nil {
		panic(info)
	}

	t.F.Seek(info.Size()-8*5, 0)
	err = binary.Read(t.F, binary.LittleEndian, &t.Tablemeta.Version)
	if err != nil {
		panic(err)
	}

	t.F.Seek(info.Size()-8*4, 0)
	err = binary.Read(t.F, binary.LittleEndian, &t.Tablemeta.Datastart)
	if err != nil {
		panic(err)
	}

	t.F.Seek(info.Size()-8*3, 0)
	err = binary.Read(t.F, binary.LittleEndian, &t.Tablemeta.Datalen)
	if err != nil {
		panic(err)
	}

	t.F.Seek(info.Size()-8*2, 0)
	err = binary.Read(t.F, binary.LittleEndian, &t.Tablemeta.Indexstart)
	if err != nil {
		panic(err)
	}

	t.F.Seek(info.Size()-8, 0)
	err = binary.Read(t.F, binary.LittleEndian, &t.Tablemeta.Indexlen)
	if err != nil {
		panic(err)
	}
}

//这里读取block的步骤:
//先从末尾构建元数据
//然后获取到position
//然后通过position获取到block

// table tree 用来管理所有的sstable
// 用链表数组来实现层级结构
type tablenode struct {
	Index   int64
	Sstable *sstable
	Next    *tablenode
}
type tabletree struct {
	Mu     sync.Mutex
	Levels []*tablenode
}

var levelMaxSize []int

func (tt *tabletree) Init(dir string) {
	con := GetConfig()
	levelMaxSize = make([]int, 10)
	levelMaxSize[0] = con.Level0Size
	levelMaxSize[1] = levelMaxSize[0] * 10
	levelMaxSize[2] = levelMaxSize[1] * 10
	levelMaxSize[3] = levelMaxSize[2] * 10
	levelMaxSize[4] = levelMaxSize[3] * 10
	levelMaxSize[5] = levelMaxSize[4] * 10
	levelMaxSize[6] = levelMaxSize[5] * 10
	levelMaxSize[7] = levelMaxSize[6] * 10
	levelMaxSize[8] = levelMaxSize[7] * 10
	levelMaxSize[9] = levelMaxSize[8] * 10

	tt.Levels = make([]*tablenode, 10)
	infos, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, info := range infos {
		if path.Ext(info.Name()) == ".db" {
			tt.LoadDBFile(path.Join(dir, info.Name()))
		}
	}
}

// 把一个sstable插入
// 返回这个sstable在当前level 链表的位置
// 注意带有锁的结构体都传引用
func (tt *tabletree) Insert(table *sstable, level int) int {
	tt.Mu.Lock()
	defer tt.Mu.Unlock()
	newnode := &tablenode{}
	newnode.Sstable = table
	p := tt.Levels[level]
	if p == nil {
		tt.Levels[level] = newnode
		newnode.Index = 0
		return 0
	} else {
		for p != nil {
			if p.Next == nil {
				newnode.Index = p.Index + 1
				p.Next = newnode
				break
			} else {
				p = p.Next
			}
		}
		return int(newnode.Index)
	}
}

// 把value插入到合适的level
func (tt *tabletree) CreateSstableAtLevel(val []Value, level int) *sstable {

	// tt.mu.Lock()
	// defer tt.mu.Unlock()

	keys := []string{}
	parsedindex := map[string]Position{}

	//数据区
	dataarea := make([]byte, 0)
	for _, v := range val {
		data, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		keys = append(keys, v.Key)
		pos := Position{}
		pos.Deleted = v.Deleted
		pos.Len = int64(len(data))
		pos.Start = int64(len(dataarea))
		parsedindex[v.Key] = pos
		dataarea = append(dataarea, data...)
	}
	sort.Strings(keys)

	//索引区
	indexarea, err := json.Marshal(parsedindex)
	if err != nil {
		panic(err)
	}

	//元数据
	meta := metaInfo{
		Version:    0,
		Datastart:  0,
		Datalen:    int64(len(dataarea)),
		Indexstart: int64(len(dataarea)),
		Indexlen:   int64(len(indexarea)),
	}

	newsstable := &sstable{
		Tablemeta:     meta,
		Parseindexmap: parsedindex,
		Sortedkeys:    keys,
	}

	index := tt.Insert(newsstable, level)
	//sstable文件以level.index.db命名
	con := GetConfig()
	filepath := con.DataDir + "/" + strconv.Itoa(level) + "." + strconv.Itoa(index) + ".db"
	//把数据写入文件
	WirteTODBFile(filepath, dataarea, indexarea, meta)

	f, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	newsstable.F = f
	newsstable.Filepath = filepath
	return newsstable
}

// 把一组value放到0层
func (tt *tabletree) CreateTable(val []Value) {
	tt.CreateSstableAtLevel(val, 0)
}

// 加载一个DB文件的内容到tabletree中
func (tt *tabletree) LoadDBFile(path string) {

	filename := filepath.Base(path)
	level, index := GetLevel(filename)

	table := &sstable{}
	//充实数据
	table.Init(path)

	//把这个sstable放到合适的位置
	newnode := &tablenode{}
	newnode.Index = int64(index)
	newnode.Sstable = table
	newnode.Next = nil

	currnode := tt.Levels[level]
	//这里不能更新newnode的index
	if currnode == nil {
		tt.Levels[level] = newnode
		return
	}
	//读取的顺序不能保证,所以如果遇到小index，直接放在前面就好了
	if currnode.Index > newnode.Index {
		newnode.Next = currnode
		tt.Levels[level] = newnode
		return
	}
	for currnode.Next != nil {
		if currnode.Next.Index > newnode.Index {
			newnode.Next = currnode.Next
			currnode.Next = newnode
			return
		} else {
			currnode = currnode.Next
		}
	}
	currnode.Next = newnode
	return
}

func (tt *tabletree) Search(key string) (Value, Status) {
	tt.Mu.Lock()
	defer tt.Mu.Unlock()
	for _, node := range tt.Levels {
		tables := make([]*sstable, 0)
		for node != nil {
			tables = append(tables, node.Sstable)
			node = node.Next
		}
		// 查找的时候要从最后一个 SSTable 开始查找
		for i := len(tables) - 1; i >= 0; i-- {
			value, searchResult := tables[i].Search(key)
			// 未找到，则查找下一个 SSTable 表
			if searchResult == SearchNone {
				continue
			} else { // 如果找到或已被删除，则返回结果
				return value, searchResult
			}
		}
	}
	return Value{}, SearchNone
}

func GetLevel(path string) (level int, index int) {
	fmt.Sscanf(path, "%d.%d.db", &level, &index)
	return
}
