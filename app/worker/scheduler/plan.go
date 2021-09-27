package scheduler

import (
	"context"
	"go-cron/app/common"
	"go-cron/app/worker/executor"
	"time"

	"github.com/gorhill/cronexpr"
)

//JobSchedulePlan 任务调度计划, 定义了要调度的任务信息和解析好的cron表达式以及下次调度时间
type JobSchedulePlan struct {
	Job      *common.Job          // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析好的cron表达式
	NextTime time.Time            // 下次调度时间
}

// 构造任务执行计划
func BuildJobSchedulePlan(job *common.Job) (*JobSchedulePlan, error) {
	var (
		expr            *cronexpr.Expression
		jobSchedulePlan *JobSchedulePlan
		err             error
	)

	// 解析JOB的cron表达式
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return nil, err
	}

	// 生成任务调度计划对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}

	//返回
	return jobSchedulePlan, nil
}

// 构造执行状态信息
func (jobSchedulePlan *JobSchedulePlan) BuildJobExecuteInfo() *executor.JobExecuteInfo {
	jobExecuteInfo := &executor.JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime, // 计算调度时间
		RealTime: time.Now(),               // 真实调度时间
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return jobExecuteInfo
}
