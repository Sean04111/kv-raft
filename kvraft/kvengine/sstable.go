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
	version int64
	//数据区起点
	datastart int64
	//数据区长度
	datalen int64
	//索引区起点
	indexstart int64
	//索引区长度
	indexlen int64
}

// 储存value的block的位置信息
// 在sstable中map和key对应
type position struct {
	Start   int64
	Len     int64
	Deleted bool
}

// sstable磁盘文件
type sstable struct {
	mu sync.Locker
	//文件句柄
	f        *os.File
	filepath string

	//元数据
	tablemeta metaInfo
	//索引
	parseindexmap map[string]position
	//排好序的key列表
	sortedkeys []string
}

func (t *sstable) Init(path string) {
	t.filepath = path
	t.mu = &sync.Mutex{}
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	t.f = f
	//load元数据和索引
	//只需要meta和index就可以推出data的区域
	//顺序不能变
	t.LoadMeta()
	t.LoadIndex()
}

func (t *sstable) LoadIndex() {
	bytes := make([]byte, t.tablemeta.indexlen)

	t.f.Seek(t.tablemeta.indexstart, 0)
	_, err := t.f.Read(bytes)
	if err != nil {
		panic(err)
	}

	t.parseindexmap = map[string]position{}

	err = json.Unmarshal(bytes, &t.parseindexmap)
	if err != nil {
		panic(err)
	}
	t.f.Seek(0, 0)

	keys := []string{}
	for k := range t.parseindexmap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	t.sortedkeys = keys

}
func (t *sstable) LoadMeta() {
	info, err := t.f.Stat()
	if err != nil {
		panic(info)
	}

	t.f.Seek(info.Size()-8*5, 0)
	err = binary.Read(t.f, binary.LittleEndian, &t.tablemeta.version)
	if err != nil {
		panic(err)
	}

	t.f.Seek(info.Size()-8*4, 0)
	err = binary.Read(t.f, binary.LittleEndian, &t.tablemeta.datastart)
	if err != nil {
		panic(err)
	}

	t.f.Seek(info.Size()-8*3, 0)
	err = binary.Read(t.f, binary.LittleEndian, &t.tablemeta.datalen)
	if err != nil {
		panic(err)
	}

	t.f.Seek(info.Size()-8*2, 0)
	err = binary.Read(t.f, binary.LittleEndian, &t.tablemeta.indexstart)
	if err != nil {
		panic(err)
	}

	t.f.Seek(info.Size()-8, 0)
	err = binary.Read(t.f, binary.LittleEndian, &t.tablemeta.datalen)
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
	index   int64
	sstable *sstable
	next    *tablenode
}
type tabletree struct {
	mu     sync.Mutex
	levels []*tablenode
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

	tt.levels = make([]*tablenode, 10)
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
	tt.mu.Lock()
	defer tt.mu.Unlock()
	newnode := &tablenode{}
	newnode.sstable = table
	p := tt.levels[level]
	index := 0
	if p == nil {
		tt.levels[level] = newnode
		newnode.index = int64(index)
		return index
	} else {
		for p.next != nil {
			p = p.next
			index++
		}
		p.next = newnode
		newnode.index = int64(index)
		return index
	}
}

// 把value插入到合适的level
func (tt *tabletree) CreateSstableAtLevel(val []Value, level int) *sstable {
	keys := []string{}
	parsedindex := map[string]position{}

	//数据区
	dataarea := make([]byte, 0)
	for _, v := range val {
		data, _ := json.Marshal(v)
		keys = append(keys, v.Key)
		pos := position{}
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
		version:    0,
		datastart:  0,
		datalen:    int64(len(dataarea)),
		indexstart: int64(len(dataarea)) + 1,
		indexlen:   int64(len(indexarea)),
	}

	newsstable := &sstable{
		tablemeta:     meta,
		parseindexmap: parsedindex,
		sortedkeys:    keys,
	}

	index := tt.Insert(newsstable, level)
	//sstable文件以level.index.db命名
	filepath := "./sstables/" + strconv.Itoa(level) + "." + strconv.Itoa(index) + ".db"
	//把数据写入文件
	WirteTODBFile(filepath, dataarea, indexarea, meta)

	f, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	newsstable.f = f
	newsstable.filepath = filepath
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
	newnode.index = int64(index)
	newnode.sstable = table
	newnode.next = nil

	currnode := tt.levels[level]
	//这里不能更新newnode的index
	if currnode == nil {
		tt.levels[level] = newnode
		return
	}
	//读取的顺序不能保证,所以如果遇到小index，直接放在前面就好了
	if currnode.index > newnode.index {
		newnode.next = currnode
		tt.levels[level] = newnode
		return
	}
	for currnode.next != nil {
		if currnode.next.index > newnode.index {
			newnode.next = currnode.next
			currnode.next = newnode
			return
		} else {
			currnode = currnode.next
		}
	}
	currnode.next = newnode
	return
}

func GetLevel(path string) (level int, index int) {
	fmt.Sscanf(path, "%d.%d.db", &level, &index)
	return
}
