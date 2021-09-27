package master

import (
	"context"
	"go-cron/app/common"

	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ApiServer struct {
	g *gin.Engine
}

func InitApiServer(cfgfile string) error {
	var (
		apiServer *ApiServer
		srv       *http.Server
		addr      string
		conf      *common.Config
		err       error
	)

	//初始化配置
	if err = common.InitConfig(cfgfile); err != nil {
		return err
	}

	//获取配置
	conf = common.GetConfig()

	//初始化jobMgr
	if err = common.InitJobMgr(); err != nil {
		return err
	}

	//初始化logMgr
	if err = common.InitLogMgr(); err != nil {
		return err
	}
	// 初始化路由
	apiServer = new(ApiServer)
	apiServer.g = gin.Default()

	InitRouter(apiServer.g)

	// 启动TCP监听
	// fmt.Println(conf.Host, conf.Port)
	addr = conf.Host + ":" + strconv.Itoa(conf.Port)
	// addr = "localhost:8099"
	srv = &http.Server{
		Addr:    addr,
		Handler: apiServer.g,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("服务监听:" + addr)

	// 优雅断开
	quit := make(chan os.Signal, 10)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
	return nil
}
