package graceful

import (
	"context"
	"errors"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application interface {
	WaitExitCommand()
}

type Graceful struct {
	MainCtx        context.Context
	Logger         *logger.Logger
	StartAction    func(ctx context.Context) error
	ShutdownAction func(ctx context.Context) error
}

func (g *Graceful) WaitExitCommand() {
	waitClosing := make(chan struct{})
	waitSignal := make(chan os.Signal, 1)
	signal.Notify(waitSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelStartAction := context.WithCancel(g.MainCtx)
	go func() {
		if err := g.StartAction(ctx); err != nil {
			var msg = "Application StartAction"
			if ctx.Err() != nil {
				g.Logger.DebugContext(ctx, msg, err)
			}
		}
		waitClosing <- struct{}{}
	}()
	g.Logger.InfoContext(ctx, "Application started")

	<-waitSignal
	cancelStartAction()
	g.Logger.InfoContext(ctx, "Application stopped")

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := g.ShutdownAction(ctx); err != nil && !errors.Is(err, ctx.Err()) {
			g.Logger.ErrorContext(ctx, "Application ShutdownAction", err)
		}
		cancel()
	}

	select {
	case <-waitClosing:
		shutdown()
	case <-time.After(3 * time.Second):
		shutdown()
	}

	g.Logger.InfoContext(ctx, "Application exited")
}
