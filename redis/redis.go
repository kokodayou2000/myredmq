package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/demdxx/gocast"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
)

type MsgEntity struct {
	MsgID string
	Key   string
	Val   string
}

var ErrNoMsg = errors.New("no msg received")

type Client struct {
	opts *ClientOptions
	pool *redis.Pool
}

// NewClient 创建一个客户端
func NewClient(network, address, password string, opts ...ClientOption) *Client {
	// 创建一个客户端 c
	c := Client{
		opts: &ClientOptions{
			network:  network,
			address:  address,
			password: password,
		},
	}
	for _, opt := range opts {
		opt(c.opts)
	}
	repairClient(c.opts)

	pool := c.getRedisPool()
	return &Client{
		pool: pool,
	}
}

func (c *Client) getRedisPool() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := c.getRedisConn()
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		MaxIdle:     c.opts.maxIdle,
		MaxActive:   c.opts.maxActive,
		IdleTimeout: time.Duration(c.opts.idleTimeoutSeconds) * time.Second,
		Wait:        c.opts.wait,
	}
}

func (c *Client) GetConn(ctx context.Context) (redis.Conn, error) {
	return c.pool.GetContext(ctx)
}

// 获取 redis 连接
func (c *Client) getRedisConn() (redis.Conn, error) {
	if c.opts.address == "" {
		panic("Cannot get redis address from config")
	}
	var diaOpts []redis.DialOption
	if len(c.opts.password) > 0 {
		diaOpts = append(diaOpts, redis.DialPassword(c.opts.password))
	}
	conn, err := redis.DialContext(
		context.Background(),
		c.opts.network,
		c.opts.address,
		diaOpts...,
	)
	if err != nil {
		return nil, err
	}
	return conn, err
}

// XADD xadd <topic> <管道缓存最大长度> <key> <val>
func (c *Client) XADD(ctx context.Context, topic string, maxLen int, key, val string) (string, error) {
	if topic == "" {
		return "", errors.New("redis XADD topic can't be empty")
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	// 默认执行的是自增 * 将自增id返回
	return redis.String(conn.Do("XADD", topic, "MAXLEN", maxLen, "*", key, val))
}

func (c *Client) XACK(ctx context.Context, topic, groupID, msgID string) error {
	if topic == "" || groupID == "" || msgID == "" {
		return errors.New("redis XACK topic | group_id | msg_id can't be empty")
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	reply, err := redis.Int64(conn.Do("XACK", topic, groupID, msgID))
	if err != nil {
		return err
	}
	if reply != 1 {
		// 非法的返回值
		return fmt.Errorf("invalid reply: %d", reply)
	}
	return nil
}

func (c *Client) XReadGroupPending(ctx context.Context, groupID, consumerID, topic string) ([]*MsgEntity, error) {
	return c.xReadGroup(ctx, groupID, consumerID, topic, 0, true)
}

func (c *Client) XReadGroup(ctx context.Context, groupID, consumerID, topic string, timeoutMilSeconds int) ([]*MsgEntity, error) {
	return c.xReadGroup(ctx, groupID, consumerID, topic, timeoutMilSeconds, false)
}

// XReadGroup group <group> <consumer> <block - ms> streams <my_streams_topic>
// 127.0.0.1:6379> xreadgroup group g1 consumer streams  t >
// 1) 1) "t"
// 2) 1) 1) "1695479204735-0"
// 2) 1) "k1"
// 2) "v1"
// 3) "k2"
// 4) "v2"
// 2) 1) "1695479272687-0"
// 2) 1) "k1"
// 2) "v1"
// 3) "k2"
// 4) "v2"
func (c *Client) xReadGroup(ctx context.Context, groupID, consumerID, topic string, timeoutMilSeconds int, pending bool) ([]*MsgEntity, error) {
	//
	if groupID == "" || consumerID == "" || topic == "" {
		return nil, errors.New("redis XREADGROUP groupID/comsumerID/topic can't be empty")
	}
	// 创建连接对象，并绑定上下文
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var rawReply interface{}
	// 执行
	if pending {
		rawReply, err = conn.Do("XREADGROUP", "GROUP", groupID, consumerID, "STREAMS", topic, "0-0")
	} else {
		rawReply, err = conn.Do("XREADGROUP", "GROUP", groupID, consumerID, "BLOCK", timeoutMilSeconds, "STREAMS", topic, ">")
	}
	if err != nil {
		return nil, err
	}
	reply, _ := rawReply.([]interface{})
	if len(reply) == 0 {
		return nil, ErrNoMsg
	}

	replyElement, _ := reply[0].([]interface{})
	if len(replyElement) != 2 {
		return nil, errors.New("invalid msg format")
	}

	var msgs []*MsgEntity
	rawMsgs, _ := replyElement[1].([]interface{})
	for _, rawMsg := range rawMsgs {
		_msg, _ := rawMsg.([]interface{})
		if len(_msg) != 2 {
			return nil, errors.New("invalid msg format")
		}
		msgID := gocast.ToString(_msg[0])
		msgBody, _ := _msg[1].([]interface{})
		if len(msgBody) != 2 {
			return nil, errors.New("invalid msg format")
		}
		msgKey := gocast.ToString(msgBody[0])
		msgVal := gocast.ToString(msgBody[1])
		msgs = append(msgs, &MsgEntity{
			MsgID: msgID,
			Key:   msgKey,
			Val:   msgVal,
		})
	}
	return msgs, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", errors.New("redis GET KEY can't be empty")
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return redis.String(conn.Do("GET", key))
}

func (c *Client) Set(ctx context.Context, key, value string) (int64, error) {
	if key == "" || value == "" {
		return -1, errors.New("redis SET key or value can't be empty")
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()
	resp, err := conn.Do("SET", key, value)
	if err != nil {
		return -1, err
	}
	if respStr, ok := resp.(string); ok && strings.ToLower(respStr) == "ok" {
		return 1, nil
	}
	return redis.Int64(resp, err)
}

func (c *Client) SetNX(ctx context.Context, key, value string) (int64, error) {
	if key == "" || value == "" {
		return -1, errors.New("redis SET key NX or value can't be empty")
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()
	resp, err := conn.Do("SET", key, value, "NX")
	if err != nil {
		return -1, err
	}
	if respStr, ok := resp.(string); ok && strings.ToLower(respStr) == "ok" {
		return 1, nil
	}
	return redis.Int64(resp, err)
}

func (c *Client) Del(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("redis DEL key can't be empty")
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return err
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return -1, errors.New("redis INCR key can't be empty")
	}
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()
	return redis.Int64(conn.Do("INCR", key))
}

// Eval EVAL "return { KEYS[1], KEYS[2], ARGV[1], ARGV[2], ARGV[3] }" 2 key1 key2 arg1 arg2 arg3
func (c *Client) Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error) {
	args := make([]interface{}, 2+len(keysAndArgs))
	args[0] = src
	args[1] = keyCount
	copy(args[2:], keysAndArgs)
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()
	return conn.Do("EVAL", args...)
}
