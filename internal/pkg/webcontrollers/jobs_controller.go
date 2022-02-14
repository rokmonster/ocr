package webcontrollers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type JobsController struct {
	Router *gin.RouterGroup
}

func (controller *JobsController) Setup() {
	// List all the jobs
	controller.Router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "jobs.html", gin.H{
			"jobs": []string{"feature not implemented", "to be added", "yet another job"},
		})
	})
}
