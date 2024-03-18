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

// Application is an interface that represents an application
type Application interface {
	// WaitExitSignal is a method that waits for the application to exit
	WaitExitSignal()
}

// Graceful is a struct that represents an application with graceful shutdown
type Graceful struct {
	// MainCtx is a context that will be used to control the application lifecycle
	MainCtx context.Context
	// Logger is a logger that will be used to log application events
	Logger *logger.Logger
	// StartAction is a function that will be called when the application starts
	StartAction func(ctx context.Context) error
	// ShutdownAction is a function that will be called when the application stops
	ShutdownAction func(ctx context.Context) error
}

func (g *Graceful) WaitExitSignal() {
	waitManualClosing := make(chan struct{})
	waitOsSignal := make(chan os.Signal, 1)
	signal.Notify(waitOsSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelMainCtx := context.WithCancel(g.MainCtx)
	go func() {
		if err := g.StartAction(ctx); err != nil {
			if ctx.Err() != nil {
				g.Logger.DebugContext(ctx, "Application can't start", err)
			}
		}
		waitManualClosing <- struct{}{}
	}()
	g.Logger.InfoContext(ctx, "Application started")

	<-waitOsSignal
	cancelMainCtx()
	g.Logger.InfoContext(ctx, "Application stopping...")

	timeout := 3 * time.Second
	shutdown := func() {
		ctx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeout)
		if err := g.ShutdownAction(ctx); err != nil && !errors.Is(err, ctx.Err()) {
			g.Logger.ErrorContext(ctx, "Application unexpected shutdown", err)
		}
		cancelShutdownTimeoutCtx()
	}

	select {
	case <-waitManualClosing:
		shutdown()
	case <-time.After(timeout):
		shutdown()
	}

	g.Logger.InfoContext(ctx, "Application exited")
}
