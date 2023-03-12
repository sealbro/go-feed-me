package main

import "github.com/sealbro/go-feed-me/pkg/graceful"

func main() {
	container := BuildApplication()

	err := container.Invoke(func(application graceful.Application) {
		application.WaitExitCommand()
	})

	if err != nil {
		panic(err)
	}
}
