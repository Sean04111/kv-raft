package kvengine

import (
	"encoding/binary"
	"os"
)

//对db文件进行操作


//把数据写入db文件
func WirteTODBFile(filepath string,dataarea,indexarea []byte,meta metaInfo){
	f,err:=os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	if err!=nil{
		panic(err)
	}
	_,err=f.Write(dataarea)
	if err!=nil{
		panic(err)
	}
	_,err = f.Write(indexarea)
	if err!=nil{
		panic(err)
	}
	binary.Write(f,binary.LittleEndian,meta.version)
	binary.Write(f,binary.LittleEndian,meta.datastart)
	binary.Write(f,binary.LittleEndian,meta.datalen)
	binary.Write(f,binary.LittleEndian,meta.indexstart)
	binary.Write(f,binary.LittleEndian,meta.indexlen)
	f.Sync()
}

//sstable获取对应的db文件的大小
func (t *sstable)GetDBSize()int64{
	info,_:=os.Stat(t.filepath)
	return info.Size()
}

//tabletree获取某一层的level的大小
func (tt *tabletree)GetLevelSize(level int)int64{
	var size int64
	currnode:=tt.levels[level]
	for currnode!=nil{
		size+=currnode.sstable.GetDBSize()
		currnode = currnode.next
	}
	return size
}
//tabletree 获取某一层的table的数量
func (tt *tabletree)GetLevelNum(level int)int{
	num:=0
	currnode:=tt.levels[level]
	for currnode!=nil{
		num++
		currnode = currnode.next
	}
	return num
}