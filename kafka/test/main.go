package main

import (
	"context"
	"encoding/json"
	"os"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/kafka"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/logger"
	redishelper "bitbucket.org/fastbanking/ring-jobscheduler-service/redis_helper"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/services"
	"github.com/gookit/ini/v2/dotenv"
	"github.com/spf13/viper"
)

func _loadenv() {
	err := dotenv.Load("./", ".env")
	if err != nil && os.Getenv("SERVICE_NAME") == "" {
		panic(err)
	}
	viper.AutomaticEnv()
}
func main() {
	_loadenv()
	logger.Configure(logger.LoggerOptions{
		ServiceName: "ring-jobscheduler-service",
		Env:         "local",
		Level:       "DEBUG",
	})

	ctx := context.Background()
	kafkaClient, err := kafka.NewKafkaClient(kafka.KOptions{
		Brokers:         viper.GetString("KAFKA_BROKERS"),
		ConsumerGroupId: viper.GetString("SERVICE_NAME") + "-CG",
		Topics:          []string{"JobSchedulerTest"},
		Tls:             true,
		ApiKey:          viper.GetString("KAFKA_API_KEY"),
		ApiSecret:       viper.GetString("KAFKA_API_SECRET"),
	})
	if err != nil {
		logger.Error(ctx, err, "Error creating Kafka client", nil)
		return
	} else {
		logger.Info(ctx, "Kafka client connected successfully", nil)
	}

	pushData(ctx, kafkaClient)
	scheduleTask(ctx, kafkaClient)
}

func pushData(ctx context.Context, client kafka.KafkaClientInterface) {
	data := `[
		{
			"task_data": [
				{
					"templateName": "sms_notification_for_bounced_nach",
					"templateModel": {
						"name": "Yadunandan",
						"amount": "25.00",
						"reason": "insufficient funds",
						"payment_link": "https://ringtest.page.link/pLtS",
						"customer_care_no": "022 41434302"
					},
					"idType": "user_id",
					"idValue": 4353760,
					"priority": 0,
					"toEmail": "YadunandanChaudhry30436@example.net",
					"fromEmail": "care@test.paywithring.com",
					"contactNo": 9326674067,
					"gcmId": "fewJnISiQEmcg8tZkIW_0O:APA91bGQbs3d5M9mcuN0KIMp5f_eVbY5aMncC0CvDOnZ3dX_GZtUlShrXafK3_gH5QVivxHTvIEwrID5O0tivugd-cbwvINiffQzeHoJ08v8P5MUV35XDUPSPLXyvYRc2tsCgUTPFUao"
				}
			],
			"task_type": "send sms",
			"task_id": "test_push_payments",
			"process_in": "1:m",
			"queue": "critical",
			"max_retry": 10
		}
	]`

	kafkaHelper := &kafka.KafkaHelperService{
		Client: client,
	}

	var dataToPush []services.TaskRequest
	if err := json.Unmarshal([]byte(data), &dataToPush); err != nil {
		logger.Error(ctx, err, "Error unmarshaling JSON data in pushData", nil)
		return
	}

	kafkaHelper.PushDataToKafka(
		ctx,
		"ring-jobscheduler-service",
		"schedule_task",
		"JobSchedulerTest",
		"",
		"",
		"Payments",
		"9dee37beda21f359a",
		dataToPush,
	)
}

func scheduleTask(ctx context.Context, client kafka.KafkaClientInterface) {
	taskService := services.NewTaskService()
	moduleService := &services.ModuleService{
		Cache: redishelper.CreateNewRedisCache(),
	}
	asyncWorker := &kafka.AsyncWorker{
		Consumer:      client,
		TaskService:   taskService,
		ModuleService: moduleService,
	}
	err := asyncWorker.Work(ctx)
	if err != nil {
		logger.Error(ctx, err, "", nil)
		return
	}
	logger.Info(ctx, "task scheduled successfully", nil)
}
