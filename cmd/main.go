package main

import (
	"log"
	"net/http"

	"github.com/edo3/minihyperproxy"
	"github.com/gorilla/mux"
)

func handleRequests(httpMux *mux.Router) {
	log.Fatal(http.ListenAndServe(":7052", httpMux))
}

func main() {
	mini := minihyperproxy.NewMinihyperProxy()
	httpMux := minihyperproxy.BuildAPI(mini)
	mini.InfoLog.Printf("Serving MiniHyperProxy on port: %v", 7052)
	handleRequests(httpMux)
}
