package cmd

import (
	"flag"
	"go-cron/app/worker"
	"os"
)

//worker子命令
func workerExecute() {
	var (
		cfgfile string
	)
	//创建worker子命令
	workerCmd := flag.NewFlagSet("worker", flag.ExitOnError)
	workerCmd.StringVar(&cfgfile, "c", "settings/worker.yml",
		//帮助信息
		"Example: go-cron worker -c ./settings/worker.yml")

	//初始化子命令
	workerCmd.Parse(os.Args[2:])
	//启动worker
	worker.Worker(cfgfile)
}
