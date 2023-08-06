package main

import "github.com/sealbro/go-feed-me/pkg/graceful"

func main() {
	err := diContainer.Invoke(func(application graceful.Application) {
		application.WaitExitCommand()
	})

	if err != nil {
		panic(err)
	}
}
