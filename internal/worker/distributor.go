package worker

import (
	"context"
	"fmt"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendEmail(ctx context.Context, payload *tasks.SendEmailPayload, opts ...asynq.Option) error
	DistributeTaskAuditLog(ctx context.Context, payload auditModel.CreateAuditLogRequest, opts ...asynq.Option) error
	DistributeTaskAuditOutboxSync(ctx context.Context, opts ...asynq.Option) error
	DistributeTaskAuditLogExport(ctx context.Context, payload auditModel.AuditLogExportPayload, opts ...asynq.Option) error
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

func (d *RedisTaskDistributor) DistributeTaskAuditLog(ctx context.Context, payload auditModel.CreateAuditLogRequest, opts ...asynq.Option) error {
	task, err := tasks.NewAuditLogCreateTask(payload)
	if err != nil {
		return fmt.Errorf("failed to create audit log task: %w", err)
	}

	info, err := d.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue audit log task: %w", err)
	}

	_ = info
	return nil
}

func (d *RedisTaskDistributor) DistributeTaskAuditOutboxSync(ctx context.Context, opts ...asynq.Option) error {
	task := tasks.NewAuditOutboxSyncTask()
	_, err := d.client.EnqueueContext(ctx, task, opts...)
	return err
}

func (d *RedisTaskDistributor) DistributeTaskAuditLogExport(ctx context.Context, payload auditModel.AuditLogExportPayload, opts ...asynq.Option) error {
	task, err := tasks.NewAuditLogExportTask(payload)
	if err != nil {
		return fmt.Errorf("failed to create audit log export task: %w", err)
	}

	_, err = d.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue audit log export task: %w", err)
	}

	return nil
}

func (d *RedisTaskDistributor) Close() error {
	return d.client.Close()
}
