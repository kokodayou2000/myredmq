package myredmq

import (
	"context"
	"myredmq/log"
	"myredmq/redis"
)

type DeadLetterMailbox interface {
	Deliver(ctx context.Context, msg *redis.MsgEntity) error
}

type DeadLetterLogger struct {
}

func NewDeadLetterLogger() *DeadLetterLogger {
	return &DeadLetterLogger{}
}

func (d *DeadLetterLogger) Deliver(ctx context.Context, msg *redis.MsgEntity) error {
	log.ErrorContext(ctx, "msg fail execedd retry limit, msg id: %s", msg.MsgID)
	return nil
}
