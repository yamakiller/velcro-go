package rds

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

var (
	_client       *redis.Client    = nil
	_sync         *redsync.Redsync = nil
	_addr         string
	_password     string = ""
	_dialTimeout  int    = 10000
	_readTimeout  int    = 10000
	_writeTimeout int    = 10000
)

var (
	ErrorPlayerIsOnline                = errors.New("player is online")
	ErrorPlayerInsufficientPermissions = errors.New("player Insufficient permissions")
	ErrorPlayerLockUnLockFail          = errors.New("player unlock fail")
	ErrorPlayerDoesNotExist            = errors.New("player does not exist")
)

const (
	playNameKey = "player_name:"
	playNickKey = "player_nick:"
	playLockKey = "player_lock:"
)

const (
	maxOfCoefficient  = 4 // 最大连接系数
	idleOfCoefficient = 2 // 空闲连接系数
)

// WithAddr 集群地址表 addr:port
func WithAddr(addr string) {
	_addr = addr
}

// WithPwd 连接验证密码
func WithPwd(pwd string) {
	_password = pwd
}

// WithDialTimeout 连接超时时间(单位毫秒)
func WithDialTimeout(millisec int) {
	_dialTimeout = millisec
}

// WithReadTimeout 读取超时时间(单位毫秒)
func WithReadTimeout(millisec int) {
	_readTimeout = millisec
}

// WithWriteTimeout 写入超时时间(单位毫秒)
func WithWriteTimeout(millisec int) {
	_writeTimeout = millisec
}

func Connection() error {

	_client = redis.NewClient(&redis.Options{
		Addr:         _addr,
		DB:           0,
		Password:     _password,
		PoolSize:     runtime.NumCPU() * maxOfCoefficient,
		MinIdleConns: runtime.NumCPU() * idleOfCoefficient,
		IdleTimeout:  time.Minute * time.Duration(15),
		DialTimeout:  time.Duration(_dialTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(_readTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(_writeTimeout) * time.Millisecond,
	})

	// 检测是否连接畅通
	if _, err := _client.Ping(context.TODO()).Result(); err != nil {
		return err
	}

	pool := goredis.NewPool(_client)
	_sync = redsync.New(pool)

	return nil
}

func Disconnect() {
	if _sync != nil {
		_sync = nil
	}

	if _client != nil {
		_client.Close()
		_client = nil
	}
}
