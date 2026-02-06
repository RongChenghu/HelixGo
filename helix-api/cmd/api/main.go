package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"helix-api/internal/app"
	"helix-api/internal/config"
)

func main() {
	// 👇 只在本地/开发环境加载
	if os.Getenv("GO_ENV") == "" || os.Getenv("GO_ENV") == "development" {
		if err := godotenv.Load(".env.development"); err != nil {
			log.Println("[warn] .env.development not loaded:", err)
		} else {
			log.Println("[info] .env.development loaded")
		}
	}
	cfg := config.LoadFromEnv()
	engine := app.Bootstrap(cfg)

	addr := ":" + cfg.AppPort
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
