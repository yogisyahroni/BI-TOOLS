package main

import (
	"insight-engine-backend/bootstrap"
	"insight-engine-backend/services"
)

func main() {
	// 1. Initialize Bootstrap
	// This is the single entry point for the entire application lifecycle
	app := bootstrap.NewApp()

	// 2. Run Application
	// This handles the server listen and graceful shutdown logic
	app.Run()

	services.LogInfo("main", "Application exited cleanly", nil)
}
