package controllers

import (
	"net/http"
	"strings"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/helpers"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/services"
	"github.com/gin-gonic/gin"
)

type TaskController struct {
	ModuleService *services.ModuleService
	TaskService   *services.TaskService
}

func (t TaskController) ScheduleTask(c *gin.Context) {
	moduleName := strings.ToLower(c.Param("module_name"))
	var requests []services.TaskRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		response := helpers.PrepareErrorResponse("invalid request", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	err := t.TaskService.EnqueueTask(c, requests, moduleName)
	if err != nil {
		response := helpers.PrepareErrorResponse(err.Error(), http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := helpers.PrepareSuccessResponse("success", http.StatusOK, "Task Scheduled Sucessfully")
		c.JSON(http.StatusOK, response)
	}
}
