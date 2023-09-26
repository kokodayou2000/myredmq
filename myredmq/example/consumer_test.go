package example

import (
	"context"
	"myredmq/myredmq"
	"myredmq/myredmq/redis"
	"testing"
	"time"
)

// 死信队列
type DemoDeadLetterMailbox struct {
	do func(msg *redis.MsgEntity)
}

func NewDemoDeadLetterMailbox(do func(msg *redis.MsgEntity)) *DemoDeadLetterMailbox {
	return &DemoDeadLetterMailbox{do: do}
}

func (d *DemoDeadLetterMailbox) Deliver(ctx context.Context, msg *redis.MsgEntity) error {
	d.do(msg)
	return nil
}

func Test_Consumer(t *testing.T) {
	client := redis.NewClient(network, address, password)

	// 接收到消息后的处理函数
	callbackFunc := func(ctx context.Context, msg *redis.MsgEntity) error {
		t.Logf("receive msg, msg id: %s, msg key: %s ,msg val: %s", msg.MsgID, msg.Key, msg.Val)
		return nil
	}
	demoDeadLetterMailbox := NewDemoDeadLetterMailbox(func(msg *redis.MsgEntity) {
		t.Logf("receive dead letter, msg id: %s, msg key: %s ,msg val: %s", msg.MsgID, msg.Key, msg.Val)
	})

	consumer, err := myredmq.NewConsumer(client,
		topic,
		consumerGroup,
		consumerID,
		callbackFunc,
		myredmq.WithMaxRetryLimit(2),
		myredmq.WithHandleMsgTimeout(2*time.Second),
		myredmq.WithDeadLetterMailbox(demoDeadLetterMailbox))
	if err != nil {
		t.Error(err)
		return
	}

	defer consumer.Stop()
	<-time.After(10 * time.Second)
}
