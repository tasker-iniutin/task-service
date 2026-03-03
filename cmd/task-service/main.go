package main

import (
	"log"

	"todo/task-service/internal/app"
)

func main() {
	a := app.CreateApp(":8080", ":50051")
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
