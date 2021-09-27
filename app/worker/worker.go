package worker

import (
	"go-cron/app/common"
	"go-cron/app/worker/executor"
	"go-cron/app/worker/logger"
	"go-cron/app/worker/register"
	"go-cron/app/worker/scheduler"
	"go-cron/app/worker/watcher"
	"log"
	"os"
	"os/signal"
)

//TODO:
//主要功能:
//1,实现watcher模块，从etcd中把job同步到内存 done
//2,实现调度模块,基于cron表达式来调度job
//3,实现执行模块，并发执行job，对job进行分布式锁，防止集群并发
//4,实现注册模块，将worker注册到master实现健康检查
//5,实现日志模块，记录日志，并记录到mongodb
func Worker(cfgfile string) {
	var (
		hb  *register.HeartBeat
		err error
	)

	//1,启动心跳
	//启动失败直接返回
	if err = register.HeartBeatInit(); err != nil {
		return
	}
	hb = register.GetHeartBeat()
	hb.KeepAlive()
	log.Println("Keepalive......")

	//2,初始化配置
	if err = common.InitConfig(cfgfile); err != nil {
		//如果失败，设置为不可用
		register.SetHealth(register.Unavailable)
	}
	log.Println("Init Config......")
	//3,初始化jobMgr
	if err = common.InitJobMgr(); err != nil {
		//如果失败，设置为不可用
		register.SetHealth(register.Unavailable)
	}
	//4,初始化logger
	logger.InitLogger()
	log.Println("Started Logger......")

	//初始化执行器executor
	executor.ExecutorInit()
	log.Println("Started Executor......")

	//初始化schduler
	scheduler.InitScheduler()
	log.Println("Started Scheduler......")

	//启动监听
	watcher.WatchJobs()
	watcher.WatchKillers()
	log.Println("Started Watcher......")

	//优雅退出
	quit := make(chan os.Signal, 10)
	signal.Notify(quit, os.Interrupt)
	<-quit
	hb.Stop()
	log.Println("Shutdown Worker ...")
}
