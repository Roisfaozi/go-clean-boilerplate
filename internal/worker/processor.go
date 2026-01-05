package worker

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
}

type RedisTaskProcessor struct {
	server *asynq.Server
	logger *logrus.Logger
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, logger *logrus.Logger) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.WithContext(ctx).Errorf("Failed to process task type %s: %v", task.Type(), err)
			}),
			Logger: NewAsynqLogger(logger),
		},
	)

	return &RedisTaskProcessor{
		server: server,
		logger: logger,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	// Register Handlers
	emailHandler := handlers.NewEmailTaskHandler(processor.logger)
	mux.HandleFunc(tasks.TypeSendEmail, emailHandler.ProcessTaskSendEmail)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}

// Adapter to make logrus compatible with asynq logger interface
type AsynqLogger struct {
	logger *logrus.Logger
}

func NewAsynqLogger(logger *logrus.Logger) *AsynqLogger {
	return &AsynqLogger{logger: logger}
}

func (l *AsynqLogger) Debug(args ...interface{}) { l.logger.Debug(args...) }
func (l *AsynqLogger) Info(args ...interface{})  { l.logger.Info(args...) }
func (l *AsynqLogger) Warn(args ...interface{})  { l.logger.Warn(args...) }
func (l *AsynqLogger) Error(args ...interface{}) { l.logger.Error(args...) }
func (l *AsynqLogger) Fatal(args ...interface{}) { l.logger.Fatal(args...) }