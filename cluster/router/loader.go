package router

import (
	"fmt"
	"os"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/yamakiller/velcro-go/cluster/proxy"
	"github.com/yamakiller/velcro-go/cluster/router/material"
	"github.com/yamakiller/velcro-go/utils/files"
	"gopkg.in/yaml.v2"
)

func Loader(filePath string, options ...RouterRpcProxyConfigOption) (*RouterGroup, error) {
	if !files.IsFileExist(filePath) {
		return nil, fmt.Errorf("file %s unfound", filePath)
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	opt := Configure(options...)

	yamlGroup := &material.RouterGroup{}
	if err := yaml.Unmarshal(b, yamlGroup); err != nil {
		return nil, err
	}

	result := &RouterGroup{routes: make([]*Router, 0), cmaps: cmap.New(), vmaps: cmap.New()}
	for _, router := range yamlGroup.Routes {
		// TODO: 构建代理对象
		p, err := proxy.NewRpcProxy(
			proxy.WithAlgorithm(opt.Algorithm),
			proxy.WithKleepalive(opt.Kleepalive),
			proxy.WithDialTimeout(opt.DialTimeout),
			proxy.WithLogger(opt.Logger),
			proxy.WithTargetHost(router.Endpoints))
		if err != nil {
			return nil, err
		}

		r := &Router{
			Proxy:    p,
			commands: make(map[string]struct{}),
		}

		// 构建角色表
		for _, rule := range router.Rules {
			r.rules |= rule
		}

		// 构建指令表
		for _, cmd := range router.Commands {
			r.commands[cmd] = struct{}{}
		}

		for _, addr := range router.Endpoints {
			result.vmaps.SetIfAbsent(addr.VAddr, r)
		}
		result.Push(r)

	}

	return result, nil
}
