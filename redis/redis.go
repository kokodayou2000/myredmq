package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
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
