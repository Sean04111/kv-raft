package kvraft

import "kv-raft/common"

// KVStateMachine
// 对kv DB的抽象，但是数据不可能一直都存在内存里面，暂时用map
// 这里后期会考虑使用LSM架构
type KVStateMachine interface {
	Get(key string) (string, common.Err)
	Put(key, value string) common.Err
	Append(key, value string) common.Err
}

type MKV struct {
	KV map[string]string
}

func NewMKV() *MKV {
	return &MKV{map[string]string{}}
}

func (kv *MKV) Get(key string) (string, common.Err) {
	value, ok := kv.KV[key]
	if !ok {
		return common.EmptyString, common.ErrNoKey
	} else {
		return value, common.OK
	}
}
func (kv *MKV) Put(key, value string) common.Err {
	kv.KV[key] = value
	return common.OK
}
func (kv *MKV) Append(key, value string) common.Err {
	kv.KV[key] += value
	return common.OK
}