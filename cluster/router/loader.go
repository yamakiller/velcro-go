package router

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/yamakiller/velcro-go/cluster/proxy"
	"github.com/yamakiller/velcro-go/cluster/router/material"
	"github.com/yamakiller/velcro-go/rpc/client/clientpool"
	"github.com/yamakiller/velcro-go/utils/files"
	"gopkg.in/yaml.v2"
)

func Loader(filePath string, algorithm string) (*RouterGroup, error) {
	if !files.IsFileExist(filePath) {
		return nil, fmt.Errorf("file %s unfound", filePath)
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	yamlGroup := &material.RouterGroup{}
	if err := yaml.Unmarshal(b, yamlGroup); err != nil {
		return nil, err
	}

	result := &RouterGroup{routes: make([]*Router, 0), cmaps: cmap.New(), vmaps: cmap.New()}
	for _, router := range yamlGroup.Routes {
		// TODO: 构建代理对象
		p, err := proxy.NewRpcProxy(
			proxy.WithAlgorithm(algorithm),
			proxy.WithPoolConfig(clientpool.IdleConfig{
				MaxIdleGlobal:      router.MaxIdleGlobal,
				MaxIdleTimeout:     time.Duration(router.MaxIdleTimeout) * time.Minute,
				MaxIdleConnTimeout: time.Duration(router.MaxIdleConnTimeout) * time.Second,
				Kleepalive:         router.Kleepalive,
			}),
			proxy.WithTargetHost(router.Endpoints))
		if err != nil {
			return nil, err
		}

		r := &Router{
			Proxy:    p,
			commands: make(map[string]int64),
		}

		// 构建角色表
		for _, rule := range router.Rules {
			r.rules |= rule
		}

		// 构建指令表
		for _, cmd := range router.Commands {
			cs := strings.Split(cmd, ",")
			if len(cs) != 2 {
				r.commands[cs[0]] = defaultTimeout
			} else {
				t, _ := strconv.ParseInt(cs[1], 10, 64)
				r.commands[cs[0]] = t
			}
		}

		for _, addr := range router.Endpoints {
			result.vmaps.SetIfAbsent(addr.VAddr, r)
		}
		result.Push(r)
	}

	return result, nil
}
