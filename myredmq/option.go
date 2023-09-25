package myredmq

import "time"

type ProducerOptions struct {
	msgQueueLen int
}

type ProducerOption func(opts *ProducerOptions)

func WithMsgQueueLen(len int) ProducerOption {
	return func(opts *ProducerOptions) {
		opts.msgQueueLen = len
	}
}

func repairProducer(opts *ProducerOptions) {
	if opts.msgQueueLen <= 0 {
		opts.msgQueueLen = 500
	}
}

type ConsumerOptions struct {
	// 每轮接收消息的超时时长
	receiveTimeout time.Duration

	// 处理消息的最大重试次数，超过此次数的时候，消息会被投递的死信队列中
	maxRetryLimit int

	// 死信队列
	deadLetterMailbox DeadLetterMailbox

	// 投递死信队列流程的超时阈值
	deadLetterDeliverTimeout time.Duration

	// 处理消息流程超时阈值
	handleMsgsTimeout time.Duration
}

type ConsumerOption func(opts *ConsumerOptions)

// WithReceiveTimeout  设定超时时间
func WithReceiveTimeout(timeout time.Duration) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.receiveTimeout = timeout
	}
}

// WithMaxRetryLimit 最大重试次数
func WithMaxRetryLimit(maxRetryLimit int) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.maxRetryLimit = maxRetryLimit
	}
}

// WithDeadLetterMailbox 死信队列
func WithDeadLetterMailbox(mailbox DeadLetterMailbox) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.deadLetterMailbox = mailbox
	}
}

func WithDeadLetterDeliverTimeout(timeout time.Duration) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.deadLetterDeliverTimeout = timeout
	}
}
func WithHandleMsgTimeout(timeout time.Duration) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.handleMsgsTimeout = timeout
	}
}

func repairConsumer(opts *ConsumerOptions) {
	if opts.receiveTimeout < 0 {
		opts.receiveTimeout = 2 * time.Second
	}
	if opts.maxRetryLimit < 0 {
		opts.maxRetryLimit = 3
	}
	if opts.deadLetterMailbox == nil {
		opts.deadLetterMailbox = NewDeadLetterLogger()
	}
	if opts.deadLetterDeliverTimeout <= 0 {
		opts.deadLetterDeliverTimeout = time.Second
	}
	if opts.handleMsgsTimeout <= 0 {
		opts.handleMsgsTimeout = time.Second
	}
}
