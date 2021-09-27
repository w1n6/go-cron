package master

import (
	"mime"

	"github.com/gin-gonic/gin"
)

func InitRouter(g *gin.Engine) *gin.RouterGroup {
	r := g.Group("")      //root根目录
	StaticFileRouter(r)   //	/static
	RegisterJobsRouter(r) //	/jobs
	RegisterListRouter(r) // 列表路由
	return r
}

func RegisterJobsRouter(r *gin.RouterGroup) {
	jobs := r.Group("/jobs")
	{
		jobs.GET("/:key", HandleGetJob)
		jobs.PUT("/:key", HandleSaveJob)
		jobs.DELETE("/:key", HandleDelJob)
		RegisterKillerRouter(jobs)
	}
}

func RegisterKillerRouter(r *gin.RouterGroup) {
	killer := r.Group("/kill")
	{
		killer.PUT("/:key", HandleKillJob)
	}
}
func RegisterListRouter(r *gin.RouterGroup) {
	r.GET("/joblist", HandleListJob)
	r.GET("/loglist", HandleListLog)
	r.GET("workerlist", HandelListWorker)
}

func StaticFileRouter(r *gin.RouterGroup) {
	_ = mime.AddExtensionType(".js", "application/javascript")
	r.Static("/static", "./static")
	r.Static("/form-generator", "./static/form-generator")
}
