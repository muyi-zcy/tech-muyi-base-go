package nacos

import (
	"github.com/muyi-zcy/tech-muyi-base-go/config"
)

// InitWithError 初始化 Nacos 并返回错误（供 starter 记录日志）
func InitWithError(cfg *config.Config) (Registry, error) {
	if cfg == nil || !cfg.Plugins.Nacos.Enabled {
		globalRegistry = &noopRegistry{}
		return globalRegistry, nil
	}
	reg, err := newNacosRegistry(cfg)
	if err != nil {
		globalRegistry = &noopRegistry{}
		return globalRegistry, err
	}
	globalRegistry = reg
	return globalRegistry, nil
}
