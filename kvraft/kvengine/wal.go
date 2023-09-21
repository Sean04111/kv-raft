package kvengine

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path"
	"sync"
)

type wal struct {
	f *os.File
	path string
	sync.Locker
}
//初始化并且load
func(wal *wal)Init(dir string)*BST{
	path:=path.Join(dir,"wal.log")
	newf,err:=os.OpenFile(path,os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
	if err!=nil{
		panic(err)
	}
	wal.path = path
	wal.f = newf
	wal.Locker = &sync.Mutex{}
	return wal.Load()
}
//写
func (wal *wal)Write(kv Value)Status{
	wal.Locker.Lock()
	defer wal.Locker.Unlock()


	data,_:=json.Marshal(kv)
	//这里先写入data的长度
	err:=binary.Write(wal.f,binary.LittleEndian,int64(len(data)))
	if err!=nil{
		return WriteFailed
	}
	err = binary.Write(wal.f,binary.LittleEndian,data)
	if err!=nil{
		return WriteFailed
	}
	return WriteSuccess
}
//把wal文件中的数据load到bst中
func (wal *wal)Load()*BST{
	wal.Locker.Lock()
	defer wal.Locker.Unlock()

	bst:=NewBST()

	//获取wal文件的信息
	info,_:=os.Stat(wal.path)
	size:=info.Size()

	//如果wal文件是空的
	if size==0{
		return bst
	}


	//把读头标到文件开头
	_,err:=wal.f.Seek(0,0)
	if err!=nil{
		panic("failed to open wal.log")
	}
	//最后把标头摆到文件末尾，为了下一次追加
	defer wal.f.Seek(size-1,0)	

	data:=make([]byte,size)
	_,err=wal.f.Read(data)
	if err!=nil{
		panic("failed to read wal.log")
	}

	datalen:=int64(0)//8 byte
	index:=int64(0)

	for index<size{
		//获取数据长度
		indexData:=data[index:index+8]
		buf:=bytes.NewBuffer(indexData)

		er:=binary.Read(buf,binary.LittleEndian,&datalen)
		if er!=nil{
			panic(er)
		}
		//获取数据
		index+=8
		valdata:=data[index:(index+datalen)]
		var val Value
		e:=json.Unmarshal(valdata,&val)
		if e!=nil{
			panic(e)
		}
		//把数据放到bst中

		if val.Deleted{
			bst.Delete(val.Key)
		}else{
			bst.Set(val)
		}
		index = index+datalen
	}
	return bst
}
//清空wal文件,前提是已经load了
func (wal *wal)Reset(){
	wal.Locker.Lock()
	wal.Locker.Unlock()
	err:=wal.f.Close()
	if err!=nil{
		panic(err)
	}
	err=os.Remove(wal.path)
	if err!=nil{
		panic(err)
	}

	newf,e:=os.OpenFile(wal.path,os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
	if e!=nil{
		panic(e)
	}
	wal.f = newf
}