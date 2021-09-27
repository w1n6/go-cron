package executor

import (
	"go-cron/app/common"
	"go-cron/app/worker/logger"
	"math/rand"

	"os/exec"
	"time"
)

// 任务执行器
type Executor struct {
	JobExecutingTable map[string]*JobExecuteInfo // 任务执行表
	JobExecuteChan    chan *JobExecuteInfo       //任务执行队列
	JobResultChan     chan *JobExecuteResult     // 任务结果队列
}

//全局任务执行器
var e *Executor

//初始化
func ExecutorInit() {
	e = &Executor{
		JobExecutingTable: make(map[string]*JobExecuteInfo),
		JobExecuteChan:    make(chan *JobExecuteInfo, 1000),
		JobResultChan:     make(chan *JobExecuteResult, 1000),
	}
	go e.ExecuteLoop()
	go e.RecordLoop()
}

//获取全局唯一Executor
func GetExecutor() *Executor {
	if e == nil {
		ExecutorInit()
		return e
	}
	return e
}

//执行协程
func (e *Executor) ExecuteLoop() {
	for info := range e.JobExecuteChan {
		go e.ExecuteJob(info)
	}
}

//记录日志协程
func (e *Executor) RecordLoop() {
	var (
		res *JobExecuteResult
		lg  *logger.Logger
		log *common.JobLog
	)
	lg = logger.GetLogger()
	for res = range e.JobResultChan {
		log = res.BuildJobLog()
		lg.CommitChan <- log
	}
}

//执行任务
func (e *Executor) ExecuteJob(info *JobExecuteInfo) {
	var (
		isExecuting bool
	)
	//为防止执行中的任务再次被调度到执行
	//如果执行信息在执行表中, 说明已经在执行了
	if _, isExecuting = e.JobExecutingTable[info.Job.Name]; isExecuting {
		return
	}
	//否则加入执行表
	e.JobExecutingTable[info.Job.Name] = info

	//执行
	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			output  []byte
			result  *JobExecuteResult
			jobMgr  *common.JobMgr
			jobLock *JobLock
		)

		// 任务结果
		result = &JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}
		//获取全局jobMgr
		jobMgr = common.GetJobMgr()

		//创建锁
		jobLock = CreateLock(jobMgr, info.Job.Name)

		// 记录任务开始时间
		result.StartTime = time.Now()

		// 随机睡眠(0~1s)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		//退出时释放锁
		defer jobLock.Unlock()
		if err = jobLock.Lock(); err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {

			// 执行shell命令
			cmd = exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)

			// 执行并捕获输出
			output, err = cmd.CombinedOutput()

			// 记录任务结束时间
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}
		// 任务执行完成后，把执行的结果返回给executor
		e.JobResultChan <- result
		delete(e.JobExecutingTable, result.ExecuteInfo.Job.Name)
	}()
}
