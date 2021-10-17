package main

import (
	"log"
	"net/http"

	"github.com/adiazny/easy-strava-upload/internal/pkg/api"
)

func main() {
	log.Println("Starting Easy-Strava-Upload Application...")

	log.Fatal(http.ListenAndServe(":8090", api.NewServer()))

}
