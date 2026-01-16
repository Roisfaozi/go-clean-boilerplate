package worker

import (
	"context"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendEmail(ctx context.Context, payload *tasks.SendEmailPayload, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}

func (d *RedisTaskDistributor) DistributeTaskSendEmail(ctx context.Context, payload *tasks.SendEmailPayload, opts ...asynq.Option) error {
	task, err := tasks.NewSendEmailTask(payload.To, payload.Subject, payload.Body)
	if err != nil {
		return fmt.Errorf("failed to create email task: %w", err)
	}

	info, err := d.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue email task: %w", err)
	}

	_ = info
	return nil
}

func (d *RedisTaskDistributor) Close() error {
	return d.client.Close()
}
