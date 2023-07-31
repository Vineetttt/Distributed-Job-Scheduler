package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/helpers"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

type TaskRequest struct {
	TaskData  interface{} `json:"task_data"`
	TaskType  string      `json:"task_type"`
	TaskID    string      `json:"task_id"`
	ProcessIn string      `json:"process_in,omitempty"`
	ProcessAt string      `json:"process_at,omitempty"`
	Queue     string      `json:"queue"`
	MaxRetry  int         `json:"max_retry"`
	Timeout   string      `json:"timeout,omitempty"`
	Deadline  string      `json:"deadline,omitempty"`
	Retention string      `json:"retention"`
}

type GeneralTask struct {
	TaskData interface{}
	TaskType string
	TaskID   string
	Options  []asynq.Option
}

type TaskService struct {
	asynqClient *asynq.Client
}

func (t TaskService) ScheduleTask(taskType string, task GeneralTask) error {
	client := t.asynqClient
	data, err := json.Marshal(task.TaskData)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if _, err := client.Enqueue(asynq.NewTask(taskType, data, task.Options...)); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (t TaskService) buildTaskOptions(request TaskRequest, moduleName string) ([]asynq.Option, error) {
	var options []asynq.Option

	if request.ProcessIn != "" && request.ProcessAt == "" {
		duration, err := helpers.ParseDuration(request.ProcessIn)
		if err != nil {
			err = fmt.Errorf("error for process_in- %v", err)
			return nil, err
		}
		options = append(options, asynq.ProcessIn(duration))
	}

	if request.ProcessAt != "" && request.ProcessIn == "" {
		processAt, err := time.Parse(time.RFC3339, request.ProcessAt)
		if err != nil {
			err = fmt.Errorf("error for process_at- %v", err)
			return nil, err
		}
		options = append(options, asynq.ProcessAt(processAt))
	}

	if request.TaskID != "" {
		options = append(options, asynq.TaskID(request.TaskID))
	}

	if request.TaskID == "" {
		err := fmt.Errorf("task_id is a required field-")
		return nil, err
	}

	if request.Queue != "" {
		queue := strings.ToLower(request.Queue)
		if queue == "low" || queue == "critical" {
			mod := strings.ToLower(moduleName)
			queueName := mod + ":" + request.Queue
			options = append(options, asynq.Queue(queueName))
		} else {
			return nil, fmt.Errorf("invalid queue- %s", request.Queue)
		}
	}

	if request.MaxRetry > 0 {
		options = append(options, asynq.MaxRetry(request.MaxRetry))
	}

	if request.Timeout != "" {
		duration, err := helpers.ParseDuration(request.Timeout)
		if err != nil {
			err = fmt.Errorf("error for timeout- %v", err)
			return nil, err
		}
		options = append(options, asynq.Timeout(duration))
	}

	if request.Deadline != "" {
		deadline, err := time.Parse(time.RFC3339, request.Deadline)
		if err != nil {
			err = fmt.Errorf("error for deadline- %v", err)
			return nil, err
		}
		options = append(options, asynq.Deadline(deadline))
	}

	if request.Retention != "" {
		duration, err := helpers.ParseDuration(request.Retention)
		if err != nil {
			err = fmt.Errorf("error for retention- %v", err)
			return nil, err
		}
		options = append(options, asynq.Retention(duration))
	} else {
		defaultRetention, err := helpers.ParseDuration("192:h")
		if err != nil {
			err = fmt.Errorf("error for default retention- %v", err)
			return nil, err
		}
		//fmt.Println("Setting default retention to:", defaultRetention)
		options = append(options, asynq.Retention(defaultRetention))
	}
	return options, nil
}

func NewTaskService() *TaskService {
	return &TaskService{
		asynqClient: asynq.NewClient(asynq.RedisClientOpt{
			Addr: viper.GetString("REDIS_URL"),
		}),
	}
}

func (t TaskService) EnqueueTask(c context.Context, requests []TaskRequest, moduleName string) error {
	var tasks []GeneralTask
	// build the tasks
	for _, req := range requests {
		taskData := req.TaskData
		options, err := t.buildTaskOptions(req, moduleName)
		if err != nil {
			return fmt.Errorf("failed to Build Task TaskID: %s Error: %s", req.TaskID, err.Error())
		}
		task := GeneralTask{
			TaskData: taskData,
			TaskType: req.TaskType,
			TaskID:   req.TaskID,
			Options:  options,
		}
		tasks = append(tasks, task)
	}
	// schedule the successfully built tasks
	for _, task := range tasks {
		if err := t.ScheduleTask(task.TaskType, task); err != nil {
			return fmt.Errorf("failed to Build Task TaskID: %s Error: %s", task.TaskID, err.Error())
		}
	}
	return nil
}
