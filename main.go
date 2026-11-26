package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/shapnesht/sharding/internal/config"
	"github.com/shapnesht/sharding/internal/handler"
	"github.com/shapnesht/sharding/internal/shard"
	"github.com/shapnesht/sharding/internal/store"
)

func main() {
	fmt.Println("Starting the server....")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Unable to fetch configs")
	}

	fmt.Println("Connecting to DB....")
	conn1, err1 := shard.Connect(context.Background(), cfg.Host, cfg.DBName, cfg.Username, cfg.Password, 5433)
	conn2, err2 := shard.Connect(context.Background(), cfg.Host, cfg.DBName, cfg.Username, cfg.Password, 5434)
	conn3, err3 := shard.Connect(context.Background(), cfg.Host, cfg.DBName, cfg.Username, cfg.Password, 5435)
	conn4, err4 := shard.Connect(context.Background(), cfg.Host, cfg.DBName, cfg.Username, cfg.Password, 5436)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		log.Fatal("Unable to connect to DB...")
	}
	fmt.Println("Connected...")

	connections := map[string]*pgx.Conn{
		"5433": conn1,
		"5434": conn2,
		"5435": conn3,
		"5436": conn4,
	}
	defer connections["5433"].Close(context.Background())
	defer connections["5434"].Close(context.Background())
	defer connections["5435"].Close(context.Background())
	defer connections["5436"].Close(context.Background())

	c := shard.NewRing([]string{"5433", "5434", "5435", "5436"})

	log.Fatal(http.ListenAndServe(":8080", handler.New(c, store.New(connections))))
}
