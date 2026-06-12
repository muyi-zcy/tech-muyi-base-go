package callloop

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/core"
	pb "github.com/muyi-zcy/tech-muyi-base-go/example/proto/echo"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"go.uber.org/zap"
)

// Config 持续互调配置
type Config struct {
	Self         string
	PeerKey      string
	Interval     time.Duration
	InitialDelay time.Duration
	Timeout      time.Duration
}

// Status 最近一次互调状态
type Status struct {
	Self      string `json:"self"`
	PeerKey   string `json:"peerKey"`
	Round     int64  `json:"round"`
	LastReply string `json:"lastReply,omitempty"`
	LastError string `json:"lastError,omitempty"`
	LastAt    int64  `json:"lastAt"`
	Running   bool   `json:"running"`
}

var (
	statusRunning atomic.Bool
	statusRound   atomic.Int64
	statusLastAt  atomic.Int64
	statusReply   atomic.Value
	statusErr     atomic.Value
	statusSelf    string
	statusPeerKey string
)

// Start 启动后台持续调用对端 gRPC Echo
func Start(starter *core.Starter, cfg Config) {
	if !starter.GetRpcManager().Enabled() {
		myLogger.Warn("RPC 未启用，跳过持续互调", zap.String("self", cfg.Self))
		return
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.InitialDelay <= 0 {
		cfg.InitialDelay = 8 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}

	statusSelf = cfg.Self
	statusPeerKey = cfg.PeerKey
	statusRunning.Store(true)

	go runLoop(starter, cfg)
	myLogger.Info("持续互调已启动",
		zap.String("self", cfg.Self),
		zap.String("peer", cfg.PeerKey),
		zap.Duration("interval", cfg.Interval),
		zap.Duration("initialDelay", cfg.InitialDelay),
	)
}

// GetStatus 获取互调状态（供 HTTP 查询）
func GetStatus() Status {
	reply, _ := statusReply.Load().(string)
	errMsg, _ := statusErr.Load().(string)
	return Status{
		Self:      statusSelf,
		PeerKey:   statusPeerKey,
		Round:     statusRound.Load(),
		LastReply: reply,
		LastError: errMsg,
		LastAt:    statusLastAt.Load(),
		Running:   statusRunning.Load(),
	}
}

func runLoop(starter *core.Starter, cfg Config) {
	time.Sleep(cfg.InitialDelay)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	round := int64(0)
	for {
		round++
		callPeer(starter, cfg, round)
		<-ticker.C
	}
}

func callPeer(starter *core.Starter, cfg Config, round int64) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	conn, err := starter.GetRpcManager().Client().GetConn(cfg.PeerKey)
	if err != nil {
		recordCall(cfg, round, "", err)
		return
	}

	message := fmt.Sprintf("%s -> %s #%d", cfg.Self, cfg.PeerKey, round)
	reply, err := pb.NewEchoServiceClient(conn).Echo(ctx, &pb.EchoRequest{Message: message})
	if err != nil {
		recordCall(cfg, round, "", err)
		return
	}

	recordCall(cfg, round, reply.GetMessage(), nil)
}

func recordCall(cfg Config, round int64, reply string, err error) {
	statusRound.Store(round)
	statusLastAt.Store(time.Now().Unix())

	if err != nil {
		statusErr.Store(err.Error())
		statusReply.Store("")
		myLogger.Warn("互调失败",
			zap.String("self", cfg.Self),
			zap.String("peer", cfg.PeerKey),
			zap.Int64("round", round),
			zap.Error(err),
		)
		return
	}

	statusErr.Store("")
	statusReply.Store(reply)
	myLogger.Info("互调成功",
		zap.String("self", cfg.Self),
		zap.String("peer", cfg.PeerKey),
		zap.Int64("round", round),
		zap.String("reply", reply),
	)
}
