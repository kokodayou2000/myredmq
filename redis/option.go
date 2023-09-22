package redis

const (
	// DefaultIdleTimeoutSeconds 默认连接池存活时间
	DefaultIdleTimeoutSeconds = 10
	// DefaultMaxActive 默认最大激活连接数
	DefaultMaxActive = 100
	// DefaultMaxIdle 默认最大空闲连接数
	DefaultMaxIdle = 20
)

type ClientOptions struct {
	maxIdle            int
	idleTimeoutSeconds int
	maxActive          int
	wait               bool
	network            string
	address            string
	password           string
}

type ClientOption func(c *ClientOptions)

func repairClient(c *ClientOptions) {
	if c.maxIdle < 0 {
		c.maxIdle = DefaultMaxIdle
	}
	if c.idleTimeoutSeconds < 0 {
		c.idleTimeoutSeconds = DefaultIdleTimeoutSeconds
	}
	if c.maxActive < 0 {
		c.maxActive = DefaultMaxActive
	}
}
