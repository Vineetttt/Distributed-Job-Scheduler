package logger

import (
	"context"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

// ElasticConfig is used for initialising elk hook for logs
//
// We haven't supported all the possible properties of elk configuration. You can extend this list
// by checking other properties available at https://github.com/elastic/go-elasticsearch/blob/main/elasticsearch.go
type ElasticConfig struct {
	Addresses    []string // A list of Elasticsearch nodes to use.
	Username     string   // Username for HTTP Basic Authentication.
	Password     string   // Password for HTTP Basic Authentication.
	CloudID      string   // Endpoint for the Elastic Service (https://elastic.co/cloud).
	APIKey       string   // Base64-encoded token for authorization; if set, overrides username/password and service token.
	ServiceToken string   // Service token for authorization; if set, overrides username/password.
}

type LoggerOptions struct {
	ServiceName string
	Env         string
	Elk         struct {
		Enable    bool
		IndexName func() string
		Config    ElasticConfig
	}
	NotifySentry bool
	Level        string
}

type customLogger struct {
	cl          *logrus.Logger
	env         string
	serviceName string
	commitId    string
}

var logger *customLogger

func setLogLevel(l string) {
	switch l {
	case "DEBUG":
		logger.cl.SetLevel(logrus.DebugLevel)
	case "TRACE":
		logger.cl.SetLevel(logrus.TraceLevel)
	default:
		logger.cl.SetLevel(logrus.InfoLevel)
	}
}

func Info(ctx context.Context, msg string, fields map[string]interface{}) {
	logger.cl.WithFields(enhanceFields(ctx, fields)).Info(msg)
}

func Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	logger.cl.WithFields(enhanceFields(ctx, fields)).Debug(msg)
}

func Error(ctx context.Context, err error, msg string, fields map[string]interface{}) {
	if err != nil {
		msg = msg + " : " + err.Error()
	}
	logger.cl.WithFields(enhanceFields(ctx, fields)).WithError(err).Error(msg)
}

func Fatal(ctx context.Context, err error, msg string, fields map[string]interface{}) {
	if err != nil {
		msg = msg + " : " + err.Error()
	}
	logger.cl.WithFields(enhanceFields(ctx, fields)).WithError(err).Fatal(msg)
}

func enhanceFields(ctx context.Context, fields map[string]interface{}) logrus.Fields {
	if fields == nil {
		fields = make(logrus.Fields)
	}
	fields["service_name"] = logger.serviceName
	fields["env"] = logger.env
	fields["revision"] = logger.commitId
	return fields
}

func Configure(cfg LoggerOptions) {
	logger = &customLogger{
		cl:          logrus.New(),
		env:         cfg.Env,
		serviceName: cfg.ServiceName,
		commitId: func() string {
			if info, ok := debug.ReadBuildInfo(); ok {
				for _, setting := range info.Settings {
					if setting.Key == "vcs.revision" {
						return setting.Value
					}
				}
			}

			return ""
		}(),
	}
	if logger.env == "local" {
		logger.cl.SetFormatter(&logrus.TextFormatter{})
	} else {
		logger.cl.SetFormatter(&logrus.JSONFormatter{})
	}

	setLogLevel(cfg.Level)

	// overriding default os.Exit func, otherwise logrus will kill the application when calling logger.Fatal
	logger.cl.ExitFunc = func(i int) {}
}
