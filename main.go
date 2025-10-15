package main

import (
	"github.com/muyi-zcy/tech-muyi-base-go/core"
)

func main() {

	// 从配置创建并初始化启动器（自动创建并初始化应用实例）
	starter, err := core.Initialize()
	if err != nil {
		panic(err)
	}
	// 启动应用（自动注册日志中间件和健康检查）
	if err := starter.Run(); err != nil {
		panic(err)
	}
}
