package cmd

import (
	"flag"
	"go-cron/app/master"

	"os"
)

//master子命令
func masterExecute() error {
	var (
		Cfgfile string
		err     error
	)
	//创建master子命令
	masterCmd := flag.NewFlagSet("master", flag.ExitOnError)
	masterCmd.StringVar(&Cfgfile, "c", "settings/master.yml",
		//帮助信息
		"Example: go-cron master -c ./settings/master.yml")

	//初始化子命令
	masterCmd.Parse(os.Args[2:])
	//初始化apiserver
	if err = master.InitApiServer(Cfgfile); err != nil {
		return err
	}
	return nil
}
