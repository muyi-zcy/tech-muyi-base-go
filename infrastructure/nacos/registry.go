package nacos

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type nacosRegistry struct {
	cfg          config.NacosConfig
	appVersion   string
	httpPort     int
	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
	registered   bool
	instanceIP   string
	instancePort uint64
	mu           sync.Mutex
}

func newNacosRegistry(appCfg *config.Config) (*nacosRegistry, error) {
	nacosCfg := appCfg.Plugins.Nacos
	if nacosCfg.ServiceName == "" {
		nacosCfg.ServiceName = appCfg.AppName
	}
	if nacosCfg.Group == "" {
		nacosCfg.Group = "XI_PLATFORM"
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(parseHost(nacosCfg.ServerAddr), parsePort(nacosCfg.ServerAddr)),
	}
	cc := constant.ClientConfig{
		NamespaceId:         nacosCfg.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "logs/nacos",
		CacheDir:            "cache/nacos",
		LogLevel:            "warn",
		Username:            nacosCfg.Username,
		Password:            nacosCfg.Password,
	}

	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "创建 Nacos NamingClient 失败")
	}

	reg := &nacosRegistry{
		cfg:          nacosCfg,
		appVersion:   appCfg.Version,
		httpPort:     appCfg.Server.Port,
		namingClient: namingClient,
	}

	if nacosCfg.ConfigEnabled && nacosCfg.ConfigDataId != "" {
		configClient, err := clients.NewConfigClient(
			vo.NacosClientParam{
				ClientConfig:  &cc,
				ServerConfigs: sc,
			},
		)
		if err != nil {
			myLogger.Warn("Nacos ConfigClient 创建失败，跳过配置中心", zap.Error(err))
		} else {
			reg.configClient = configClient
			if err := reg.loadRemoteConfig(); err != nil {
				myLogger.Warn("拉取 Nacos 远程配置失败", zap.Error(err))
			}
		}
	}

	myLogger.Info("Nacos 插件已启用",
		zap.String("serviceName", nacosCfg.ServiceName),
		zap.String("serverAddr", nacosCfg.ServerAddr),
	)
	return reg, nil
}

func (r *nacosRegistry) Enabled() bool { return true }

func (r *nacosRegistry) Register(_ context.Context, grpcPort int, metadata map[string]string) error {
	ip, err := localIP()
	if err != nil {
		return err
	}
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["version"] = r.appVersion
	metadata["protocol"] = "grpc"
	metadata["language"] = "go"
	if r.httpPort > 0 {
		metadata["httpPort"] = strconv.Itoa(r.httpPort)
	}

	param := vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        uint64(grpcPort),
		ServiceName: r.cfg.ServiceName,
		GroupName:   r.cfg.Group,
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metadata,
	}
	ok, err := r.namingClient.RegisterInstance(param)
	if err != nil {
		return errors.Wrap(err, "Nacos 注册实例失败")
	}
	if !ok {
		return errors.New("Nacos 注册实例返回 false")
	}

	r.mu.Lock()
	r.registered = true
	r.instanceIP = ip
	r.instancePort = uint64(grpcPort)
	r.mu.Unlock()

	myLogger.Info("Nacos 注册成功",
		zap.String("service", r.cfg.ServiceName),
		zap.String("ip", ip),
		zap.Int("grpcPort", grpcPort),
	)
	return nil
}

func (r *nacosRegistry) Deregister(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.registered {
		return nil
	}
	param := vo.DeregisterInstanceParam{
		Ip:          r.instanceIP,
		Port:        r.instancePort,
		ServiceName: r.cfg.ServiceName,
		GroupName:   r.cfg.Group,
		Ephemeral:   true,
	}
	ok, err := r.namingClient.DeregisterInstance(param)
	if err != nil {
		return errors.Wrap(err, "Nacos 注销实例失败")
	}
	if !ok {
		myLogger.Warn("Nacos 注销实例返回 false")
	}
	r.registered = false
	myLogger.Info("Nacos 实例已注销", zap.String("service", r.cfg.ServiceName))
	return nil
}

func (r *nacosRegistry) NamingClient() NamingClient {
	return &nacosNamingAdapter{client: r.namingClient, group: r.cfg.Group}
}

func (r *nacosRegistry) loadRemoteConfig() error {
	content, err := r.configClient.GetConfig(vo.ConfigParam{
		DataId: r.cfg.ConfigDataId,
		Group:  r.cfg.Group,
	})
	if err != nil {
		return err
	}
	if content == "" {
		return nil
	}
	if err := viper.MergeConfig(strings.NewReader(content)); err != nil {
		return err
	}
	myLogger.Info("Nacos 远程配置已合并", zap.String("dataId", r.cfg.ConfigDataId))

	return r.configClient.ListenConfig(vo.ConfigParam{
		DataId: r.cfg.ConfigDataId,
		Group:  r.cfg.Group,
		OnChange: func(_, _, dataId, data string) {
			if data == "" {
				return
			}
			if err := viper.MergeConfig(strings.NewReader(data)); err != nil {
				myLogger.Warn("Nacos 配置热更新失败", zap.Error(err))
				return
			}
			myLogger.Info("Nacos 配置已热更新", zap.String("dataId", dataId))
		},
	})
}

type nacosNamingAdapter struct {
	client naming_client.INamingClient
	group  string
}

func (a *nacosNamingAdapter) Subscribe(serviceName, groupName string, callback func(instances []ServiceInstance)) error {
	if groupName == "" {
		groupName = a.group
	}
	return a.client.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		GroupName:   groupName,
		SubscribeCallback: func(services []model.Instance, err error) {
			if err != nil {
				myLogger.Warn("Nacos 订阅回调错误", zap.Error(err))
				return
			}
			callback(convertInstances(services))
		},
	})
}

func (a *nacosNamingAdapter) Unsubscribe(serviceName, groupName string) error {
	if groupName == "" {
		groupName = a.group
	}
	return a.client.Unsubscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		GroupName:   groupName,
	})
}

func (a *nacosNamingAdapter) SelectInstances(serviceName, groupName string) ([]ServiceInstance, error) {
	if groupName == "" {
		groupName = a.group
	}
	instances, err := a.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   groupName,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}
	return convertInstances(instances), nil
}

func convertInstances(instances []model.Instance) []ServiceInstance {
	result := make([]ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		result = append(result, ServiceInstance{
			IP:       inst.Ip,
			Port:     inst.Port,
			Weight:   inst.Weight,
			Healthy:  inst.Healthy,
			Metadata: inst.Metadata,
		})
	}
	return result
}

func parseHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func parsePort(addr string) uint64 {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 8848
	}
	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return 8848
	}
	return port
}

func localIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "127.0.0.1", nil
	}
	return hostname, nil
}
