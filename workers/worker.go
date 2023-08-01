package main

import (
	"log"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/workers/tasks"
	"github.com/hibiken/asynq"
)

func main() {
	redisConnection := asynq.RedisClientOpt{
		Addr: "localhost:6379",
	}

	worker := asynq.NewServer(redisConnection, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"payments:critical": 1,
			"payments:low":      1,
		},
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(
		"send sms",
		tasks.Send_SMS_Worker,
	)
	if err := worker.Run(mux); err != nil {
		log.Fatal(err)
	}
}
