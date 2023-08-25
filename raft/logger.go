package raft



//系统日志的设计还需要进一步设计，这里实现了用es远程收集和zap本地磁盘文件写入的方法，
//es远程收集太慢了，会影响系统的正确性
//暂时用zap文件写入替代
import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const EsAddr string = "http://43.143.238.162:9200/"
const mapping = `
{
  "mappings": {
    "properties": {
	"time":{
		"type":"keyword"
	}
    "act":{
		"type":"keyword"
	 }
	 "event":{
		"type":"keyword"
	 } 
    }
  }
}`

type EsLog struct {
	Time  string `json:"time"`
	Event string `json:"event"`
	Act   string `json:"act"`
}

type EsLogger struct {
	EsC   *elastic.Client
	Index string
}

// 初始化es日志器



func InitEsLogger(index int) (EsLogger, error) {
	c, e := elastic.NewClient(
		elastic.SetURL(EsAddr),
		elastic.SetSniff(false),
	)
	if e != nil {
		return EsLogger{}, e
	} else {
		return EsLogger{EsC: c, Index: strconv.Itoa(index)}, nil
	}
}

// 为节点创建es日志表
func (this *EsLogger) CreateIndex() error {
	if ok, _ := this.EsC.IndexExists(this.Index).Do(context.Background()); ok {
		return errors.New("index already exsits !")
	}
	_, e := this.EsC.CreateIndex(this.Index).BodyJson(mapping).Do(context.Background())
	if e != nil {
		return e
	}
	return nil
}

// 写入es日志 act:事件类型 event:事件内容
func (this *EsLogger) Insert(act string, event string) error {
	log := EsLog{}
	log.Act = act
	log.Event = event
	log.Time = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	_, e := this.EsC.Index().Index(this.Index).BodyJson(log).Do(context.Background())
	return e
}

// 清理本次测试的信息（待定）
func (this *EsLogger) Clean() {
	this.EsC.DeleteIndex(this.Index).Do(context.Background())
}

func (rf *Raft) LoggerInit() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	filename := "./logs/" + strconv.Itoa(rf.me) + ".log"
	file, _ := os.Create(filename)
	filewriter := zapcore.AddSync(file)
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(config)
	core := zapcore.NewCore(encoder, filewriter, zapcore.DebugLevel)
	rf.logger = zap.New(core, zap.AddCaller()).Sugar()
}
