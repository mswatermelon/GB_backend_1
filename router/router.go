package router

import (
	"fmt"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

func main() {
	// define mux.Router
	router := chi.NewRouter()
	// register anonymous handler function for root path for GET method
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "GET HANDLER")
		if err != nil {
			return
		}
	})
	// register anonymous handler function for root path for POST method
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "POST HANDLER")
		if err != nil {
			return
		}
	})
	// run the server passing as a router our chi router object
	log.Fatal(http.ListenAndServe(":8080", router))
}
