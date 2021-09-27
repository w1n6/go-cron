package executor

import (
	"context"
	"go-cron/app/common"
	"time"
)

//JobExecuteInfo 执行状态信息,定义了执行job信息,理论上的调度时间,实际的调度时间以及用于取消command执行的cancel函数
type JobExecuteInfo struct {
	Job        *common.Job        //任务信息
	PlanTime   time.Time          // 理论上的调度时间
	RealTime   time.Time          // 实际的调度时间
	CancelCtx  context.Context    // 任务command的context
	CancelFunc context.CancelFunc //  用于取消command执行的cancel函数
}

//JobExecuteResult 任务执行结果 定义了任务的执行结果, 执行状态, 脚本输出, 脚本错误原因, 启动时间, 结束时间
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo // 执行状态
	Output      []byte          // 脚本输出
	Err         error           // 脚本错误原因
	StartTime   time.Time       // 启动时间
	EndTime     time.Time       // 结束时间
}

func (res *JobExecuteResult) BuildJobLog() *common.JobLog {
	jobLog := &common.JobLog{
		JobName:      res.ExecuteInfo.Job.Name,
		Command:      res.ExecuteInfo.Job.Command,
		Output:       string(res.Output),
		PlanTime:     res.ExecuteInfo.PlanTime.Format(common.TimeLayout),
		ScheduleTime: res.ExecuteInfo.RealTime.Format(common.TimeLayout),
		StartTime:    res.StartTime.Format(common.TimeLayout),
		EndTime:      res.EndTime.Format(common.TimeLayout),
	}
	if res.Err == nil {
		jobLog.Err = ""
	} else {
		jobLog.Err = res.Err.Error()
	}
	return jobLog
}
