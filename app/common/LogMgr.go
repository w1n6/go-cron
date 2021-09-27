package common

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongodb日志管理 保存有一个mongo客户端连接 以及一个collection
type LogMgr struct {
	Client        *mongo.Client
	LogDB         *mongo.Database
	LogCollection *mongo.Collection
}

// 任务日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

var lm *LogMgr //全局日志管理器

func InitLogMgr() error {
	var (
		conf       *Config
		client     *mongo.Client
		db         *mongo.Database
		collection *mongo.Collection
		err        error
	)

	//获取config
	conf = GetConfig()

	//连接mongo
	if client, err = mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(conf.Uri),
		options.Client().SetConnectTimeout(time.Duration(conf.ConnectTimeout)*time.Millisecond)); err != nil {
		return err
	}

	//选择数据库
	db = client.Database(conf.Database)

	//选择collection
	collection = db.Collection(conf.Collection)
	lm = &LogMgr{
		Client:        client,
		LogDB:         db,
		LogCollection: collection,
	}
	return nil
}

func GetLogMgr() *LogMgr {
	if lm == nil {
		InitLogMgr()
		return lm
	}
	return lm
}

//查看任务日志
func (logMgr *LogMgr) ListLog(name string, skip int, limit int) ([]*JobLog, error) {
	var (
		fi     *JobLogFilter
		opts   *options.FindOptions
		cursor *mongo.Cursor
		jobLog *JobLog
		logs   []*JobLog
		err    error
	)

	// 过滤条件
	fi = &JobLogFilter{JobName: name}

	//查询条件
	opts = options.Find()

	//按时间逆序
	opts.SetSort(bson.D{primitive.E{Key: "startTime", Value: -1}})

	//设置skip及limit
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	//查询
	if cursor, err = logMgr.LogCollection.Find(context.TODO(), fi, opts); err != nil {
		return nil, err
	}
	// 延迟释放游标
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		jobLog = &JobLog{}

		// 反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue // 有日志不合法 跳过
		}
		logs = append(logs, jobLog)
	}
	return logs, nil
}
