package app

import (
	"context"

	"github.com/letianxing/volta-brain/internal/config"
	"github.com/letianxing/volta-brain/internal/infra/natsio"
	"github.com/letianxing/volta-brain/internal/runtime"
)

// App 封装真实部署场景下的 runtime 与 NATS 客户端。
type App struct {
	runtime *runtime.Runtime
}

// New 创建应用实例。
func New(cfg config.Config) (*App, error) {
	client, err := natsio.NewClient(natsio.Config{
		URL:  cfg.NATSURL,
		Name: cfg.ServiceName,
	})
	if err != nil {
		return nil, err
	}

	return &App{
		runtime: runtime.New(cfg, client),
	}, nil
}

// Start 启动应用。
func (a *App) Start(ctx context.Context) error {
	return a.runtime.Start(ctx)
}

// Close 关闭应用。
func (a *App) Close() error {
	return a.runtime.Close()
}
