package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/helpers"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/logger"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/services"
)

type Event struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	ModuleName string
	Password   string
	Data       []services.TaskRequest `json:"data"`
}

type AsyncWorker struct {
	Consumer      KafkaClientInterface `di.inject:"kafkaClient"`
	TaskService   *services.TaskService
	ModuleService *services.ModuleService
}

func (a *AsyncWorker) Work(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "Asynq work was done", nil)
			a.Consumer.Close()
			return nil
		default:
			logger.Info(ctx, "Fetching from kafka records", nil)
			fetches := a.Consumer.Poll(ctx)
			logger.Info(ctx, "Start Record Iter", nil)
			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()
				err := a.JobIdentifier(ctx, record.Value)
				if err != nil {
					return err
				}
			}
		}
	}
}

func (a *AsyncWorker) JobIdentifier(ctx context.Context, payload []byte) error {
	logger.Info(ctx, "Start Job Identifier", nil)
	var jevent Event
	err := json.Unmarshal(payload, &jevent)
	if err != nil {
		logger.Error(ctx, err, "Error unmarshaling payload", nil)
		return err
	}
	logger.Info(ctx, "Processing async job for event:"+jevent.Name, nil)

	if jevent.Name == "schedule_task" {
		_, dbPassword, err := a.ModuleService.GetModuleCredentials(jevent.ModuleName)
		if err != nil {
			logger.Error(ctx, err, "", nil)
		}

		match := helpers.ComparePassword(dbPassword, jevent.Password)
		if match {
			a.TaskService.EnqueueTask(ctx, jevent.Data, jevent.ModuleName)
		} else {
			logger.Error(ctx, fmt.Errorf("authentication failed"), "", nil)
		}
	}
	return nil
}
