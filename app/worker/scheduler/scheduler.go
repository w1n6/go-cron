package scheduler

import (
	"go-cron/app/worker/executor"
	"time"
)

//Scheduler		调度器，保存了任务事件队列和任务调度计划表，执行状态表，以及执行结果队列
type Scheduler struct {
	JobEventChan chan *JobEvent              // 任务事件队列
	JobPlanTable map[string]*JobSchedulePlan // 任务调度计划表
}

var sc *Scheduler //全局单例调度器

//初始化调度器
func InitScheduler() {
	sc = &Scheduler{
		JobEventChan: make(chan *JobEvent, 1000),
		JobPlanTable: make(map[string]*JobSchedulePlan),
	}
	go sc.ScheduleLoop()
}

//返回全局调度器
func GetScheduler() *Scheduler {
	if sc == nil {
		InitScheduler()
		return sc
	}
	return sc
}

// 调度当前任务表所有任务，返回下次应该调度的时间
func (sc *Scheduler) Schedule() time.Duration {
	var (
		jobPlan       *JobSchedulePlan
		now           time.Time
		isExecuting   bool
		nearTime      *time.Time
		scheduleAfter time.Duration
		e             *executor.Executor
	)

	// 如果任务表为空，下次1s后调度
	if len(sc.JobPlanTable) == 0 {
		return 1 * time.Second
	}

	//获取执行器
	e = executor.GetExecutor()

	// 当前时间
	now = time.Now()

	// 遍历所有任务
	for _, jobPlan = range sc.JobPlanTable {
		//如果任务在执行就跳过调度
		if _, isExecuting = e.JobExecutingTable[jobPlan.Job.Name]; isExecuting {
			continue
		}
		//如果job下次执行的时间早于或相等当前时间
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {

			//构造执行信息并发送给执行器
			e.JobExecuteChan <- jobPlan.BuildJobExecuteInfo()

			// 更新下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		// 统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	// 下次调度间隔（最近要执行的任务调度时间 - 当前时间）
	scheduleAfter = (*nearTime).Sub(now)

	//返回下次应该调度的时间
	return scheduleAfter
}

// 处理任务事件
func (sc *Scheduler) handleJobEvent(jobEvent *JobEvent) {
	var (
		jobSchedulePlan *JobSchedulePlan
		jobExecuteInfo  *executor.JobExecuteInfo
		jobExecuting    bool
		jobExisted      bool
		err             error
	)

	//判断任务事件类型
	switch jobEvent.EventType {
	// 保存任务事件
	case JOB_EVENT_SAVE:
		//构造任务调度计划
		if jobSchedulePlan, err = BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		//存到任务调度计划表
		sc.JobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	// 删除任务事件
	case JOB_EVENT_DELETE:
		//如果存在于任务调度计划表则删除
		if _, jobExisted = sc.JobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(sc.JobPlanTable, jobEvent.Job.Name)
		}
	//强杀任务事件
	case JOB_EVENT_KILL:
		if jobExecuteInfo, jobExecuting =
			executor.GetExecutor().JobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			//取消shell进程
			jobExecuteInfo.CancelFunc()
		}
	}
}

//调度循环
func (sc *Scheduler) ScheduleLoop() {
	var (
		jobEvent      *JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
	)
	// 初始化一次(1秒)
	scheduleAfter = sc.Schedule()

	// 调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	//调度任务
	for {
		//监听
		select {
		//有任务变化
		case jobEvent = <-sc.JobEventChan:
			//处理jobEvent事件
			sc.handleJobEvent(jobEvent)

		//最近的任务到期了
		case <-scheduleTimer.C:
		}
		// 调度一次任务
		scheduleAfter = sc.Schedule()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}
