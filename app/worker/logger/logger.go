package logger

import (
	"context"
	"go-cron/app/common"
	"log"
	"time"
)

//Logger记录日志并保存到mongdb,如果日志记录达到MaxQueueSize，则同步到mongdb
type Logger struct {
	LogMgr     *common.LogMgr      //日志管理器，负责与mongdb通信
	CommitChan chan *common.JobLog //提交日志消息队列
	logBatch   []interface{}       //日志批次保存
	MaxLog     int                 //最大保存日志条数
}

var lg *Logger

//初始化日志管理器
func InitLogger() {
	var (
		logMgr *common.LogMgr
		conf   *common.Config
	)
	//获取配置
	conf = common.GetConfig()
	//获取日志管理器
	logMgr = common.GetLogMgr()
	lg = &Logger{
		LogMgr:     logMgr,
		CommitChan: make(chan *common.JobLog, 1000),
		MaxLog:     conf.MaxLog,
	}
	go lg.SyncLoop()
}

//获取日志管理器
func GetLogger() *Logger {
	if lg == nil {
		InitLogger()
		return lg
	}
	return lg
}

//提交日志协程
func (lg *Logger) SyncLoop() {
	var (
		jobLog *common.JobLog
		timer  time.Ticker
	)
	timer = *time.NewTicker(5 * time.Minute)
	for {
		select {
		//如果接收到日志
		case jobLog = <-lg.CommitChan:
			//先把日志追加到batch中
			lg.logBatch = append(lg.logBatch, jobLog)
			log.Println(*jobLog)
			//如果超过设置最大日志数，则同步
			if len(lg.logBatch) > lg.MaxLog {
				go lg.Sync()
			}
		//超过5min 自动同步
		case <-timer.C:
			go lg.Sync()
		}
	}
}

func (lg *Logger) Sync() error {
	if _, err := lg.LogMgr.LogCollection.InsertMany(context.TODO(), lg.logBatch); err != nil {
		return err
	}
	return nil
}
