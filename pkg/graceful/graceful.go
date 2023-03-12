package graceful

import (
	"context"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application interface {
	WaitExitCommand()
}

type Graceful struct {
	Logger         *logger.Logger
	StartAction    func(ctx context.Context) error
	ShutdownAction func(ctx context.Context) error
}

func (g *Graceful) WaitExitCommand() {
	waitClosing := make(chan struct{})
	waitSignal := make(chan os.Signal, 1)
	signal.Notify(waitSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelStartAction := context.WithCancel(context.Background())
	go func() {
		if err := g.StartAction(ctx); err != nil {
			var msg = "Application StartAction"
			if ctx.Err() != nil {
				g.Logger.Ctx(ctx).Debug(msg, zap.Error(err))
			} else {
				g.Logger.Ctx(ctx).Fatal(msg, zap.Error(err))
			}
		}
		waitClosing <- struct{}{}
	}()
	g.Logger.Ctx(ctx).Info("Application started")

	<-waitSignal
	cancelStartAction()
	g.Logger.Ctx(ctx).Info("Application stopped")

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := g.ShutdownAction(ctx); err != nil && err != ctx.Err() {
			g.Logger.Ctx(ctx).Error("Application ShutdownAction", zap.Error(err))
		}
		cancel()
	}

	select {
	case <-waitClosing:
		shutdown()
	case <-time.After(3 * time.Second):
		shutdown()
	}

	g.Logger.Ctx(ctx).Info("Application exited")
}
