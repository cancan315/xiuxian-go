package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/redis"
	"xiuxian/server-go/internal/http/router"
)

func main() {
	_ = godotenv.Load()

	if err := db.Init(); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	if err := redis.Init(); err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 注册路由
	router.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	log.Printf("Go server is running on %s", addr)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to run server: %v", err)
	}
}
