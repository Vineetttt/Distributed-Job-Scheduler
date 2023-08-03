package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/services"
	"github.com/twmb/franz-go/pkg/kgo"
)

const ServiceName = "job-scheduler_service"
const (
	ApiCallEvent           = "send-sms_api_call_event"
	SettlementStatusUpdate = "settlement_status_update"
	// schema subject
	SettlementStatusSchema = "settlement_status_update-api_call"
)

type KafkaHelperService struct {
	Client KafkaClientInterface `di.inject:"kafkaClient"`
}
type KafkaPayload struct {
	Namespace  string
	Name       string
	ModuleName string
	Password   string
	Data       interface{}
}

// PushDataToKafka pushes the data to kafka by creating the payload and initialising all the headers
// whatever data that needs to be sent can be sent in data interface field with own go struct type
// event name can be passed in eventName similarly for serviceName,topicName
func (k *KafkaHelperService) PushDataToKafka(ctx context.Context, serviceName, eventName, topicName, schemaSubject, key string, module string, pass string, data []services.TaskRequest) {
	namespace := "testing"
	payload := KafkaPayload{
		Namespace:  fmt.Sprintf("%s.%s", namespace, serviceName),
		Name:       eventName,
		ModuleName: module,
		Password:   pass,
		Data:       data,
	}
	headerMap := []kgo.RecordHeader{
		{Key: "Content-Type", Value: []byte("application/json")},
		{Key: "Schema-Url", Value: []byte("")},
		{Key: "Schema-Subject", Value: []byte(schemaSubject)},
		{Key: "Schema-Version", Value: []byte("1")},
	}
	jsnPayload, _ := json.Marshal(payload)
	k.Client.Produce(ctx, &kgo.Record{
		Key:     []byte(key),
		Topic:   topicName,
		Value:   jsnPayload,
		Headers: headerMap,
	})
}
