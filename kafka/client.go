package kafka

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"sync"
	"time"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/logger"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
)

type KafkaClientInterface interface {
	Close()
	Produce(ctx context.Context, record *kgo.Record)
	Poll(ctx context.Context) kgo.Fetches
}
type KafkaClient struct {
	//franz-go client of kafka
	kcl *kgo.Client
	//wait group used by kafka producer for async produce
	producerWaitGroup *sync.WaitGroup
}
type KOptions struct {
	Brokers         string
	ConsumerGroupId string
	Topics          []string
	Tls             bool
	ApiKey          string
	ApiSecret       string
}

// var client KafkaAdapterInterface
// Initialize new Kafka client
func NewKafkaClient(options KOptions) (KafkaClientInterface, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(options.Brokers, ",")...),
		kgo.ConsumerGroup(options.ConsumerGroupId),
		kgo.ConsumeTopics(options.Topics...),
		kgo.ProduceRequestTimeout(1 * time.Second),
	}
	// add sasl options if options.ApiKey or options.ApiSecret is not nil
	if options.ApiKey != "" && options.ApiSecret != "" {
		plain_mechanism := plain.Auth{
			User: options.ApiKey,
			Pass: options.ApiSecret,
		}
		opts = append(opts, kgo.SASL(plain_mechanism.AsMechanism()))
	}
	if options.Tls {
		tlsDialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: 10 * time.Second}}
		opts = append(opts, kgo.Dialer(tlsDialer.DialContext))
	}
	r, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return &KafkaClient{
		kcl:               r,
		producerWaitGroup: &sync.WaitGroup{},
	}, nil
}
func (k *KafkaClient) Produce(ctx context.Context, record *kgo.Record) {
	logger.Info(ctx, "kafka produce initiated", nil)
	k.producerWaitGroup.Add(1)
	k.kcl.Produce(ctx, record, func(rec *kgo.Record, err error) {
		defer k.producerWaitGroup.Done()
		if err != nil {
			logger.Fatal(ctx, err, err.Error(), nil)
		}
	})
	k.producerWaitGroup.Wait()
}
func (k KafkaClient) Poll(ctx context.Context) kgo.Fetches {
	return k.kcl.PollRecords(ctx, 1)
}
func (k *KafkaClient) Close() {
	k.kcl.Close()
}
