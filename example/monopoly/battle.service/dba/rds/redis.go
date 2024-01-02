package rds

import (
	"context"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

var (
	client       *redis.Client    = nil
	sync         *redsync.Redsync = nil
	address         string
	password     string = ""
	dialTimeout  int    = 10000
	readTimeout  int    = 10000
	writeTimeout int    = 10000
)

const (
	maxOfCoefficient  = 4 // 最大连接系数
	idleOfCoefficient = 2 // 空闲连接系数
)

// WithAddr 集群地址表 addr:port
func WithAddr(addr string) {
	address = addr
}

// WithPwd 连接验证密码
func WithPwd(pwd string) {
	password = pwd
}

// WithDialTimeout 连接超时时间(单位毫秒)
func WithDialTimeout(millisec int) {
	dialTimeout = millisec
}

// WithReadTimeout 读取超时时间(单位毫秒)
func WithReadTimeout(millisec int) {
	readTimeout = millisec
}

// WithWriteTimeout 写入超时时间(单位毫秒)
func WithWriteTimeout(millisec int) {
	writeTimeout = millisec
}

func Connection() error {

	client = redis.NewClient(&redis.Options{
		Addr:         address,
		DB:           0,
		Password:     password,
		PoolSize:     runtime.NumCPU() * maxOfCoefficient,
		MinIdleConns: runtime.NumCPU() * idleOfCoefficient,
		IdleTimeout:  time.Minute * time.Duration(15),
		DialTimeout:  time.Duration(dialTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(readTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(writeTimeout) * time.Millisecond,
	})

	// 检测是否连接畅通
	if _, err := client.Ping(context.TODO()).Result(); err != nil {
		return err
	}

	pool := goredis.NewPool(client)
	sync = redsync.New(pool)

	return nil
}

func Disconnect() {
	if sync != nil {
		sync = nil
	}

	if client != nil {
		client.Close()
		client = nil
	}
}
