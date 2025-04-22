package main

import (
	api "ninja-agent/api"
	// bot "ninja-agent/bot"
	// "os"
)

func main() {

	// mode := os.Getenv("MODE")
	// switch mode {
	// case "bot":
	// 	bot.RunBot()
	// case "api":
	// 	api.RunAPI()
	// }

	api.RunAPI()
}
