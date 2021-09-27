package common

import (
	"encoding/json"
)

//一些共有的类 与操作数据库etcd mongodb相关

//job类
type Job struct {
	Name     string `json:"name"`     //任务名字,唯一标识，不应重复
	Command  string `json:"command"`  //任务shell命令，默认使用bash解释器
	CronExpr string `json:"cronExpr"` //cron表达式, * * * * * * * , (支持 7 位，精确到秒、分、时、日、月、周、年), eg: */5 * * * * * * ,意为每5s执行一次
}

//解包成job []byte->job
func UnPack(data []byte) (*Job, error) {
	var (
		err error
		job *Job
	)
	job = &Job{}
	if err = json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	return job, nil
}
