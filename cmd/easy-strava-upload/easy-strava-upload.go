package main

import (
	"log"
	"net/http"

	"github.com/adiazny/easy-strava-upload/internal/pkg/api"
	"github.com/rs/cors"
)

func main() {
	log.Println("Starting Easy-Strava-Upload Application...")

	log.Fatal(http.ListenAndServe(":8090", cors.Default().Handler(api.NewServer())))

}
