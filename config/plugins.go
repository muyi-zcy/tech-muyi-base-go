package config

// PluginsConfig 可插拔基础设施配置
type PluginsConfig struct {
	Nacos NacosConfig `mapstructure:"nacos"`
	Rpc   RpcConfig   `mapstructure:"rpc"`
}

// NacosConfig Nacos 注册与配置中心
type NacosConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	ServerAddr    string `mapstructure:"serverAddr"`
	Namespace     string `mapstructure:"namespace"`
	Group         string `mapstructure:"group"`
	ServiceName   string `mapstructure:"serviceName"`
	ConfigEnabled bool   `mapstructure:"configEnabled"`
	ConfigDataId  string `mapstructure:"configDataId"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
}

// RpcConfig gRPC 服务暴露与消费
type RpcConfig struct {
	Enabled  bool              `mapstructure:"enabled"`
	Protocol string            `mapstructure:"protocol"`
	Registry string            `mapstructure:"registry"`
	Server   RpcServerConfig   `mapstructure:"server"`
	Client   RpcClientConfig   `mapstructure:"client"`
	Static   map[string]string `mapstructure:"static"`
}

// RpcServerConfig gRPC Server 配置
type RpcServerConfig struct {
	Port             int  `mapstructure:"port"`
	MaxRecvMsgSize   int  `mapstructure:"maxRecvMsgSize"`
	EnableReflection bool `mapstructure:"enableReflection"`
}

// RpcClientConfig gRPC Client 配置
type RpcClientConfig struct {
	DefaultTimeoutMs int               `mapstructure:"defaultTimeoutMs"`
	MaxRetry         int               `mapstructure:"maxRetry"`
	Services         map[string]string `mapstructure:"services"`
}

// GetPluginsConfig 获取插件配置
func GetPluginsConfig() PluginsConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if GlobalConfig != nil {
		return GlobalConfig.Plugins
	}
	return PluginsConfig{}
}

// GetNacosConfig 获取 Nacos 配置
func GetNacosConfig() NacosConfig {
	return GetPluginsConfig().Nacos
}

// GetRpcConfig 获取 RPC 配置
func GetRpcConfig() RpcConfig {
	return GetPluginsConfig().Rpc
}
