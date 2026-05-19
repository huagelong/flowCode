package main

import (
	"log"

	"anserflow/internal/bootstrap"
)

func main() {
	cfg, err := bootstrap.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		log.Fatalf("bootstrap app: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
