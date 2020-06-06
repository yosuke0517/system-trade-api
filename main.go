package main

import (
	"app/bitflyer"
	"app/config"
	"app/utils"
	"fmt"
	"log"
	"net/http"
)

func handler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Hello World, %s!!", request.URL.Path[1:])
}

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	log.Println("test")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)

	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	fmt.Println(apiClient)
}
