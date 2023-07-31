package api

import (
	"net/http"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/auth"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/controllers"
	redishelper "bitbucket.org/fastbanking/ring-jobscheduler-service/redis_helper"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/services"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
)

func Routes() *gin.Engine {
	// Bootstrapping our services
	cache := redishelper.CreateNewRedisCache()
	authMiddleware := &auth.Middleware{
		Cache: *cache,
	}
	moduleService := services.ModuleService{
		Cache: cache,
	}
	taskService := services.NewTaskService()
	moduleController := &controllers.ModuleController{
		ModuleService: &moduleService,
	}
	taskController := &controllers.TaskController{
		ModuleService: &moduleService,
		TaskService:   taskService,
	}

	// Defining gin.Engine
	router := gin.Default()

	router.GET("/monitoring/*any", gin.WrapH(func() *http.ServeMux {
		dashboard := asynqmon.New(asynqmon.Options{
			// RootPath specifies the root for asynqmon app
			// RedisConnOpt specifies the Redis connection options for the asynqmon app,
			RootPath:     "/monitoring",
			RedisConnOpt: asynq.RedisClientOpt{Addr: "localhost:6379"},
		})

		server := http.NewServeMux()
		server.Handle(dashboard.RootPath()+"/", dashboard)
		return server
	}()))

	api := router.Group("/api/v1")

	api.POST("/modules", moduleController.AddModule)
	api.POST("/modules/:module_name/tasks/schedule", authMiddleware.AuthMiddleware, taskController.ScheduleTask)

	return router
}
