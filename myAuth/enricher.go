package myAuth

import (
	"context"
	"fmt"
	"sync"
)

// SessionEnricher 在 Session 加载后注入 extras，不写入平台 context。
type SessionEnricher func(ctx context.Context, sess *Session) error

var (
	enricherMu sync.RWMutex
	enrichers  = make(map[string]SessionEnricher)
)

// RegisterSessionEnricher 注册命名 SessionEnricher。
func RegisterSessionEnricher(name string, fn SessionEnricher) {
	if name == "" || fn == nil {
		return
	}
	enricherMu.Lock()
	defer enricherMu.Unlock()
	enrichers[name] = fn
}

func runEnrichers(ctx context.Context, sess *Session, names []string) error {
	if len(names) == 0 {
		return nil
	}
	enricherMu.RLock()
	defer enricherMu.RUnlock()
	for _, name := range names {
		fn, ok := enrichers[name]
		if !ok {
			return fmt.Errorf("myAuth: session enricher %q not registered", name)
		}
		if err := fn(ctx, sess); err != nil {
			return err
		}
	}
	return nil
}

func allEnricherNames() []string {
	enricherMu.RLock()
	defer enricherMu.RUnlock()
	names := make([]string, 0, len(enrichers))
	for name := range enrichers {
		names = append(names, name)
	}
	return names
}
