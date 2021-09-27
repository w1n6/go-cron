package master

import (
	"fmt"
	"go-cron/app/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

//GET 获取job接口
// GET  /jobs/:key

func HandleGetJob(c *gin.Context) {
	var (
		err    error
		key    string
		jobMgr *common.JobMgr

		job *common.Job
	)
	//获取url参数
	key = c.Param("key")

	//获取全局job管理器
	jobMgr = common.GetJobMgr()

	//获取job
	if job, err = jobMgr.GetJob(common.JobSaveDir + key); err != nil {
		c.JSON(http.StatusBadRequest, common.HttpResponse{
			HasErr: common.RequestFail,
			Msg:    err.Error(),
			Data:   nil,
		})
	}

	//如果成功返回job
	c.JSON(http.StatusOK, common.HttpResponse{
		HasErr: common.RequestSuccess,
		Msg:    "Sucess",
		Data:   job,
	})
}

//PUT 添加job接口
//PUT  /jobs/:key
func HandleSaveJob(c *gin.Context) {
	var (
		err    error
		key    string
		jobMgr *common.JobMgr
		job    common.Job
		prejob *common.Job
	)
	//获取url参数
	key = c.Param("key")
	if err = c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, common.HttpResponse{
			HasErr: common.RequestFail,
			Msg:    err.Error(),
			Data:   nil,
		})
	}

	//获取全局job管理器
	jobMgr = common.GetJobMgr()

	//保存job
	if prejob, err = jobMgr.SaveJob(common.JobSaveDir+key, &job); err != nil {
		c.JSON(http.StatusBadRequest, common.HttpResponse{
			HasErr: common.RequestFail,
			Msg:    err.Error(),
			Data:   nil,
		})
	}
	//如果成功返回被覆盖的job
	c.JSON(http.StatusOK, common.HttpResponse{
		HasErr: common.RequestSuccess,
		Msg:    "Sucess",
		Data:   prejob,
	})
}

//DELETE 删除job接口
//DELETE  /jobs/:key
func HandleDelJob(c *gin.Context) {
	var (
		err    error
		key    string
		jobMgr *common.JobMgr
		prejob *common.Job
	)

	//获取url参数
	key = c.Param("key")

	//获取全局job管理器
	jobMgr = common.GetJobMgr()

	//删除job
	if prejob, err = jobMgr.DelJob(common.JobSaveDir + key); err != nil {
		c.JSON(http.StatusBadRequest, common.HttpResponse{
			HasErr: common.RequestFail,
			Msg:    err.Error(),
			Data:   nil,
		})
	}

	//返回被删除的job
	c.JSON(http.StatusOK, common.HttpResponse{
		HasErr: common.RequestSuccess,
		Msg:    "Sucess",
		Data:   prejob,
	})
}

//LIST 获取目录下的所有job
//GET	/jobs/list

func HandleListJob(c *gin.Context) {
	var (
		err     error
		dir     string
		jobMgr  *common.JobMgr
		jobList []*common.Job
	)

	//获取全局job管理器
	jobMgr = common.GetJobMgr()

	//获取所有job
	if jobList, err = jobMgr.ListJob(dir); err != nil {
		c.JSON(http.StatusBadRequest, common.HttpResponse{
			HasErr: common.RequestFail,
			Msg:    err.Error(),
			Data:   nil,
		})
	}
	c.JSON(http.StatusOK, common.HttpResponse{
		HasErr: common.RequestSuccess,
		Msg:    "Sucess",
		Data:   jobList,
	})
}

func HandleKillJob(c *gin.Context) {
	var (
		jobMgr *common.JobMgr
		key    string
		err    error
	)
	jobMgr = common.GetJobMgr()

	//获取url参数
	key = c.Param("key")
	if err = jobMgr.KillJob(common.JobKillDir + key); err != nil {
		c.JSON(http.StatusBadRequest, common.HttpResponse{
			HasErr: common.RequestFail,
			Msg:    err.Error(),
			Data:   nil,
		})
	}
	c.JSON(http.StatusOK, common.HttpResponse{
		HasErr: common.RequestSuccess,
		Msg:    "Sucess",
		Data:   nil,
	})
}

func HandleListLog(c *gin.Context) {
	var (
		logMgr *common.LogMgr
		name   string
		skip   int
		limit  int
		logs   []*common.JobLog
		err    error
	)
	logMgr = common.GetLogMgr()

	//获取name
	name = c.Query("name")
	//skip默认0
	if skip, err = strconv.Atoi(c.DefaultQuery("skip", "0")); err != nil {
		skip = 0
	}
	//limit默认20
	if limit, err = strconv.Atoi(c.DefaultQuery("limit", "20")); err != nil {
		limit = 20
	}

	if logs, err = logMgr.ListLog(name, skip, limit); err != nil {
		fmt.Println("失败")
		c.JSON(http.StatusBadRequest, common.HttpResponse{
			HasErr: common.RequestFail,
			Msg:    err.Error(),
			Data:   nil,
		})
	}
	c.JSON(http.StatusOK, common.HttpResponse{
		HasErr: common.RequestSuccess,
		Msg:    "Sucess",
		Data:   logs,
	})
}

func HandelListWorker(c *gin.Context) {
	var (
		jobMgr     *common.JobMgr
		workerList []string
		err        error
	)

	//获取全局job管理器
	jobMgr = common.GetJobMgr()

	//获取所有job
	if workerList, err = jobMgr.GetWorkerList(); err != nil {
		c.JSON(http.StatusBadRequest, common.HttpResponse{
			HasErr: common.RequestFail,
			Msg:    err.Error(),
			Data:   nil,
		})
	}
	fmt.Println(workerList)
	c.JSON(http.StatusOK, common.HttpResponse{
		HasErr: common.RequestSuccess,
		Msg:    "Sucess",
		Data:   workerList,
	})
}
