package myredmq

import (
	"context"
	"errors"
	"fmt"
	"myredmq/log"
	"myredmq/redis"
)

type MsgCallback func(ctx context.Context, msg *redis.MsgEntity) error

type Consumer struct {
	// consumer 生命周期管理
	ctx  context.Context
	stop context.CancelFunc

	// 接收到 msg 时执行的 回调函数，由使用方法定义
	callbackFunc MsgCallback

	// redis 客户端，基于 redis 实现 message queue
	client *redis.Client

	// 消费者 topic
	topic string

	// 所需的topic
	groupID string

	// 当前节点的消费者 id
	consumerID string

	// 各消息累计失败次数
	failureCnts map[redis.MsgEntity]int

	// 一些用户自定义的配置
	opts *ConsumerOptions
}

func NewConsumer(client *redis.Client, topic, groupID, consumerID string, callbackFunc MsgCallback, opts ...ConsumerOption) (*Consumer, error) {
	ctx, stop := context.WithCancel(context.Background())
	c := Consumer{
		ctx:          ctx,
		stop:         stop,
		callbackFunc: callbackFunc,
		client:       client,
		topic:        topic,
		groupID:      groupID,
		consumerID:   consumerID,
		failureCnts:  make(map[redis.MsgEntity]int),
		opts:         &ConsumerOptions{},
	}

	if err := c.checkParam(); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(c.opts)
	}
	// 修复用户可能输入的错误数据
	repairConsumer(c.opts)

	go c.run()

	return &c, nil
}

func (c *Consumer) checkParam() error {
	if c.callbackFunc == nil {
		return errors.New("callback function can't be empty")
	}

	if c.client == nil {
		return errors.New("redis client can't be empty")
	}
	if c.topic == "" || c.consumerID == "" || c.groupID == "" {
		return errors.New("topic | group_id | consumer_id can't be empty")
	}
	return nil
}

// Stop use Context stop
func (c *Consumer) Stop() {
	c.stop()
}

func (c *Consumer) run() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		msgs, err := c.receive()
		if err != nil {
			log.ErrorContextf(c.ctx, "receive msg failed,err %v", err)
			continue
		}
		tctx, _ := context.WithTimeout(c.ctx, c.opts.handleMsgsTimeout)
		c.handlerMsgs(tctx, msgs)

		tctx, _ = context.WithTimeout(c.ctx, c.opts.deadLetterDeliverTimeout)
		c.deliverDeadLetter(tctx)

		// pending 消息接收处理
		pendingMsgs, err := c.receivePending()
		if err != nil {
			log.ErrorContextf(c.ctx, "pending msg received failed,err: %v", err)
			continue
		}
		tctx, _ = context.WithTimeout(c.ctx, c.opts.handleMsgsTimeout)
		// 处理这些消息
		c.handlerMsgs(tctx, pendingMsgs)
	}
}

// 从 group中读取数据
func (c *Consumer) receive() ([]*redis.MsgEntity, error) {
	fmt.Println(c.groupID, c.consumerID, c.topic, int(c.opts.receiveTimeout.Milliseconds()))
	msgs, err := c.client.XReadGroup(c.ctx, c.groupID, c.consumerID, c.topic, int(c.opts.receiveTimeout.Milliseconds()))
	if err != nil && !errors.Is(err, redis.ErrNoMsg) {
		return nil, err
	}
	return msgs, err
}

func (c *Consumer) receivePending() ([]*redis.MsgEntity, error) {
	pendingMsgs, err := c.client.XReadGroupPending(c.ctx, c.groupID, c.consumerID, c.topic)
	if err != nil && !errors.Is(err, redis.ErrNoMsg) {
		return nil, err
	}
	return pendingMsgs, nil
}

// 处理消息
func (c *Consumer) handlerMsgs(ctx context.Context, msgs []*redis.MsgEntity) {
	for _, msg := range msgs {
		if err := c.callbackFunc(ctx, msg); err != nil {
			c.failureCnts[*msg]++
			continue
		}
		// callback 执行成功，进行 ack
		if err := c.client.XACK(ctx, c.topic, c.groupID, msg.MsgID); err != nil {
			log.ErrorContextf(ctx, "msg ack failed, msg id: %s, err: %v", msg.MsgID, err)
			continue
		}

		delete(c.failureCnts, *msg)
	}
}

// 投递到死信队列
func (c *Consumer) deliverDeadLetter(ctx context.Context) {
	// 对于失败过多的消息，投递到死信队列中
	for msg, failureCnt := range c.failureCnts {
		if failureCnt < c.opts.maxRetryLimit {
			continue
		}
		// 投递死信队列
		if err := c.opts.deadLetterMailbox.Deliver(ctx, &msg); err != nil {
			log.ErrorContextf(c.ctx, "dead letter deliver failed, msg id: %s, err: %v", msg.MsgID, err)
		}

		// 执行 ack 响应
		if err := c.client.XACK(ctx, c.topic, c.groupID, msg.MsgID); err != nil {
			log.ErrorContextf(c.ctx, "msg ack failed, msg id: %s, err: %v", msg.MsgID, err)
			continue
		}

		// 对于ack成功的，将其从failure map中删除
		delete(c.failureCnts, msg)
	}
}
