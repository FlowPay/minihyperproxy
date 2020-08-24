package minihyperproxy

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func handleRequests(httpMux *mux.Router) {
	log.Fatal(http.ListenAndServe(":7052", httpMux))
}

func main() {
	mini := NewMinihyperProxy()
	httpMux := buildAPI(mini)
	handleRequests(httpMux)
}
