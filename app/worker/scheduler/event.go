package scheduler

import "go-cron/app/common"

const (
	JOB_EVENT_SAVE   = iota //保存事件
	JOB_EVENT_DELETE        //删除事件
	JOB_EVENT_KILL          //强杀事件
)

//JobEvent 任务变化事件, 定义了事件类型(实际上是int), 发生变化的任务信息
type JobEvent struct {
	EventType int         //事件类型 0:保存事件 1:删除事件 2:强杀事件
	Job       *common.Job //任务信息
}

// 任务变化事件有2种：1）更新任务 2）删除任务 3):强杀任务
func BuildJobEvent(eventType int, job *common.Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}
