package watcher

import (
	"go-cron/app/common"
	"go-cron/app/worker/scheduler"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

func WatchJobs() error {
	var (
		kvs        []*mvccpb.KeyValue
		kvPair     *mvccpb.KeyValue
		job        *common.Job
		jobMgr     *common.JobMgr
		watchChan  *clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName    string
		jobEvent   *scheduler.JobEvent
		err        error
	)
	//获取全局jobMgr
	jobMgr = common.GetJobMgr()

	//开始监控
	if kvs, watchChan, err = jobMgr.WatchDir(common.JobSaveDir); err != nil {
		return err
	}
	// 1,当前有哪些任务
	for _, kvPair = range kvs {
		// 反序列化json得到Job
		if job, err = common.UnPack(kvPair.Value); err == nil {
			jobEvent = scheduler.BuildJobEvent(scheduler.JOB_EVENT_SAVE, job)
			// 同步给scheduler(调度协程)
			scheduler.GetScheduler().JobEventChan <- jobEvent
		}
	}
	// 2,监听变化事件
	go func() { // 监听协程
		for watchResp = range *watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 任务保存事件
					if job, err = common.UnPack(watchEvent.Kv.Value); err != nil {
						continue
					}
					// 构建一个更新Event
					jobEvent = scheduler.BuildJobEvent(scheduler.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE: // 任务被删除了
					// Delete /cron/jobs/job10
					jobName = common.ExtractKeyName(common.JobSaveDir, string(watchEvent.Kv.Key))

					job = &common.Job{Name: jobName}

					// 构建一个删除Event
					jobEvent = scheduler.BuildJobEvent(scheduler.JOB_EVENT_DELETE, job)
				}
				// 变化推给scheduler
				scheduler.GetScheduler().JobEventChan <- jobEvent
			}
		}
	}()
	return nil
}

func WatchKillers() error {
	var (
		job        *common.Job
		jobMgr     *common.JobMgr
		watchChan  *clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent   *scheduler.JobEvent
		jobName    string
		err        error
	)
	//获取全局jobMgr
	jobMgr = common.GetJobMgr()

	//开始监控
	if _, watchChan, err = jobMgr.WatchDir(common.JobKillDir); err != nil {
		return err
	}
	// 监听变化事件
	go func() { // 监听协程
		for watchResp = range *watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 杀死任务事件
					jobName = common.ExtractKeyName(string(watchEvent.Kv.Key), common.JobKillDir)
					job = &common.Job{
						Name: jobName,
					}
					// 构建一个更新Event
					jobEvent = scheduler.BuildJobEvent(scheduler.JOB_EVENT_KILL, job)

					// 变化推给scheduler
					scheduler.GetScheduler().JobEventChan <- jobEvent
				case mvccpb.DELETE: // 任务被删除了
				}

			}
		}
	}()
	return nil
}
