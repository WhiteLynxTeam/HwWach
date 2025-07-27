package main

import "HwWach/internal/app"

func main() {
	app, err := app.NewApp()
	if err != nil {
		println(err)
		return
	}
	app.Run()
}
