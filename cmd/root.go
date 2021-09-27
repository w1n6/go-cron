package cmd

import (
	"fmt"
	"os"
)

//根命令
func Execute() {
	var err error

	//用法
	usage := `Usage:	
go-cron help	获取帮助文档
	master	以master运行
	worker	以worker运行
	`

	//参数少于2
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	// 判断子命令
	switch os.Args[1] {
	case "help":
		fmt.Println(usage) //打印帮助文档
	case "master":
		if err = masterExecute(); err != nil {
			fmt.Println(err)
		}
	case "worker":
		workerExecute()
	default:
		fmt.Println(usage)
		fmt.Println("第一个参数错误")
	}
}
