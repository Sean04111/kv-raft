package kvengine

import "sync"

type Config struct {
	// 数据目录
	DataDir string
	// 0 层的 所有 SsTable 文件大小总和的最大值，单位 MB，超过此值，该层 SsTable 将会被压缩到下一层
	Level0Size int
	// 每层中 SsTable 表数量的阈值，该层 SsTable 将会被压缩到下一层
	PartSize int
	// 内存表的 kv 最大数量，超出这个阈值，内存表将会被保存到 SsTable 中
	Threshold int
	// 压缩内存、文件的时间间隔，多久进行一次检查工作
	CheckInterval int
}

var once *sync.Once = &sync.Once{}

// 常驻内存
var config Config

// ConfigInit 初始化数据库配置
func ConfigInit(con Config) {
	once.Do(func() {
		config = con
	})
}

// GetConfig 获取数据库配置
func GetConfig() Config {
	return config
}
