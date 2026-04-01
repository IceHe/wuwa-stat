package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"wuwa/stat/backend/internal/config"
	"wuwa/stat/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	fmt.Println("正在创建数据库表...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.EnsureSchema(ctx, database); err != nil {
		log.Fatal(err)
	}

	fmt.Println("数据库表创建成功！")
	db.PrintSchemaSummary()
}
