package main

import "github.com/MatTwix/Pull-Request-Assigner/services/pr-service/internal/app"

const configPath = "./configs/config.yml"

func main() {
	app.Run(configPath)
}
