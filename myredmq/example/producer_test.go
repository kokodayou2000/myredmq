package example

import (
	"context"
	"myredmq/myredmq"
	"myredmq/myredmq/redis"
	"testing"
)

func Test_Producer(t *testing.T) {
	client := redis.NewClient(network, address, password)

	producer := myredmq.NewProducer(client, myredmq.myredmq.WithMsgQueueLen(10))
	ctx := context.Background()
	// xread  streams my_test_topic 0-0
	msgID, err := producer.SendMsg(ctx, topic, "test_k", "test_v")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(msgID)
}
