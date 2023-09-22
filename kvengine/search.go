package kvengine

import (
	"encoding/json"
)

// 关于搜索
// 现在内存表中找,如果内存表中找不到,到tabletree中找
// 这里是在sstable文件中找的步骤
func (t *sstable) Search(key string) (Value, Status) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	//使用二分查找
	l, r := 0, len(t.Sortedkeys)-1
	var ans Position
	ans.Start = -1
	for l <= r {
		mid := (l + r) >> 1
		if t.Sortedkeys[mid] == key {
			ans = t.Parseindexmap[t.Sortedkeys[mid]]
			if ans.Deleted {
				return Value{}, SearchDeleted
			}
			break
		}
		if t.Sortedkeys[mid] < key {
			l = mid + 1
		}
		if t.Sortedkeys[mid] > key {
			r = mid - 1
		}
	}

	//没找到
	if ans.Start == -1 {
		return Value{}, SearchNone
	}

	//找到了
	//从文件中拿数据
	_, err := t.F.Seek(ans.Start, 0)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, ans.Len)
	_, err = t.F.Read(buf)
	if err != nil {
		panic(err)
	}

	var data Value
	err = json.Unmarshal(buf, &data)
	if err != nil {
		panic(err)
	}
	return data, SearchSuccess
}
