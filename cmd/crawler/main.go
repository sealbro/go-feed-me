package main

func main() {
	app, err := provideApp()
	if err != nil {
		panic(err)
	}

	app.WaitExitSignal()
}
