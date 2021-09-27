package common

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

//jobmgr job管理器, 保存了一个etcd的连接, 有client, kv, lease以及watcher API子集
//实现了job的Get获取,Delete删除,Save保存以及Kill强杀
type JobMgr struct {
	Client  *clientv3.Client //客户端连接
	KV      clientv3.KV      //KV API子集
	Lease   clientv3.Lease   //Lease API子集
	Watcher clientv3.Watcher //Watcher API子集
}

var jm *JobMgr //全局任务管理器

//JobMgr全局任务管理器初始化
func InitJobMgr() error {
	var (
		conf *Config
		err  error
	)
	//初始化
	jm = new(JobMgr)

	//获取配置
	conf = GetConfig()

	//初始化etcd
	if jm.Client, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: time.Millisecond * time.Duration(conf.DialTimeout),
	}); err != nil {
		return err
	}

	//初始化KV
	jm.KV = clientv3.NewKV(jm.Client)

	//初始化lease
	jm.Lease = clientv3.NewLease(jm.Client)
	jm.Watcher = clientv3.NewWatcher(jm.Client)
	return nil
}

//获取全局jobMgr
func GetJobMgr() *JobMgr {

	if jm == nil {
		InitJobMgr()
		return jm
	}
	return jm
}

//GET
func (jobmgr *JobMgr) GetJob(key string) (*Job, error) {
	var (
		getsResp *clientv3.GetResponse
		job      *Job
		err      error
	)
	//TODO: 添加timeout ctx
	if getsResp, err = jobmgr.KV.Get(context.TODO(), key); err != nil {
		return nil, err
	}

	//如果没有获取到数据
	if len(getsResp.Kvs) == 0 {
		err = errors.New("没有获取到数据")
		return nil, err
	}

	//获取到数据 则反序列化
	if err = json.Unmarshal(getsResp.Kvs[0].Value, &job); err != nil {
		return nil, err
	}
	return job, nil
}

//SAVE
func (jobmgr *JobMgr) SaveJob(key string, job *Job) (*Job, error) {
	var (
		jobInfo []byte
		putResp *clientv3.PutResponse
		preJob  *Job
		err     error
	)

	//反序列化job信息
	if jobInfo, err = json.Marshal(job); err != nil {
		return nil, err
	}

	//修改job
	//TODO: 添加timeout ctx
	if putResp, err = jobmgr.KV.Put(context.TODO(), key, string(jobInfo),
		clientv3.WithPrevKV()); err != nil {
		return nil, err
	}

	//如果没有覆盖,也就是新增job
	if putResp.PrevKv == nil {
		return nil, nil
	}

	//如果有覆盖则反序列化之前的数据
	if err = json.Unmarshal(putResp.PrevKv.Value, &preJob); err != nil {
		err = nil
		return nil, nil
	}
	return preJob, nil
}

//DELETE
func (jobmgr *JobMgr) DelJob(key string) (*Job, error) {
	var (
		delResp *clientv3.DeleteResponse
		prejob  *Job
		err     error
	)

	//TODO: 添加timeout ctx
	if delResp, err = jobmgr.KV.Delete(context.TODO(), key,
		clientv3.WithPrevKV()); err != nil {
		return nil, err
	}

	//返回被删除的任务,即使序列化失败，也返回成功
	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &prejob); err != nil {
			err = nil
			return nil, err
		}
	}
	return prejob, nil
}

//job列表
func (jobmgr *JobMgr) ListJob(dir string) ([]*Job, error) {
	var (
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *Job
		jobList []*Job
		err     error
	)

	if getResp, err = jobmgr.KV.Get(context.TODO(), dir,
		clientv3.WithPrefix()); err != nil {
		return nil, err
	}

	if len(getResp.Kvs) == 0 {
		return nil, err
	}
	jobList = make([]*Job, 0)
	for _, kvPair = range getResp.Kvs {
		job = new(Job)
		if err = json.Unmarshal(kvPair.Value, &job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	return jobList, nil
}

//强杀job
//实际上是在/jobs/kill + key PUT 数据
func (jobmgr *JobMgr) KillJob(key string) error {
	var (
		leaseGrantResponse *clientv3.LeaseGrantResponse
		leaseID            clientv3.LeaseID
		err                error
	)

	if leaseGrantResponse, err = jobmgr.Lease.Grant(context.TODO(), 1); err != nil {
		return err
	}
	leaseID = leaseGrantResponse.ID
	if _, err = jobmgr.KV.Put(context.TODO(), key, "KILL",
		clientv3.WithLease(leaseID)); err != nil {
		return err
	}
	return nil
}

//监控目录
func (jobmgr *JobMgr) WatchDir(dir string) ([]*mvccpb.KeyValue, *clientv3.WatchChan, error) {
	var (
		getResp            *clientv3.GetResponse
		watchStartRevision int64
		watcherChan        clientv3.WatchChan
		err                error
	)
	if getResp, err = jobmgr.KV.Get(context.TODO(), dir, clientv3.WithPrefix()); err != nil {
		return nil, nil, err
	}

	//获取Revision
	watchStartRevision = getResp.Header.Revision + 1
	watcherChan = jobmgr.Watcher.Watch(context.TODO(), dir,
		clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
	return getResp.Kvs, &watcherChan, nil
}

//获取worker
func (jobmgr *JobMgr) GetWorkerList() ([]string, error) {
	var (
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		res     []string
		key     string
		err     error
	)
	if getResp, err = jobmgr.KV.Get(context.TODO(), JobWorkerDir,
		clientv3.WithPrefix()); err != nil {
		return nil, err
	}
	for _, kvPair = range getResp.Kvs {
		key = string(kvPair.Key)
		res = append(res, ExtractKeyName(key, JobWorkerDir))
	}
	return res, nil
}
