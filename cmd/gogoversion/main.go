package main

import "github.com/markitos-it/markitos-it-gogoversion/internal/app"

var version = "dev"
var run = app.Run

func main() {
	run(version)
}
