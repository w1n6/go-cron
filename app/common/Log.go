package common

//JobLog任务执行日志 存储在mongdb中，保存有任务信息，执行信息，脚本输出
type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`           // 任务名字
	Command      string `json:"command" bson:"command"`           // 脚本命令
	Err          string `json:"err" bson:"err"`                   // 错误原因
	Output       string `json:"output" bson:"output"`             // 脚本输出
	PlanTime     string `json:"planTime" bson:"planTime"`         // 计划开始时间
	ScheduleTime string `json:"scheduleTime" bson:"scheduleTime"` // 实际调度时间
	StartTime    string `json:"startTime" bson:"startTime"`       // 任务执行开始时间
	EndTime      string `json:"endTime" bson:"endTime"`           // 任务执行结束时间
}
