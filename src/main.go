package main

import (
	"net/http"
	"log"
	"fmt"
	mux2 "github.com/gorilla/mux"
	"./handlers"
)

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("URL: %v; Method: %v; Origin: %v\n", r.URL.Path, r.Method, r.Header.Get("Origin"))
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := mux2.NewRouter()

	mux.HandleFunc(`/api/user/{nick}/create`, handlers.CreateUser).Methods("POST")
	mux.HandleFunc(`/api/user/{nick}/profile`, handlers.GetUser).Methods("GET")
	mux.HandleFunc(`/api/user/{nick}/profile`, handlers.UpdateUser).Methods("POST")

	mux.HandleFunc(`/api/forum/create`, handlers.CreateForum).Methods("POST")
	mux.HandleFunc(`/api/forum/{slug}/details`, handlers.GetForum).Methods("GET")
	mux.HandleFunc(`/api/forum/{slug}/create`, handlers.CreateThread).Methods("POST")
	mux.HandleFunc(`/api/forum/{slug}/threads`, handlers.GetThreads).Methods("GET")

	mux.HandleFunc(`/api/thread/{slug_or_id}/create`, handlers.CreatePosts).Methods("POST")
	mux.HandleFunc(`/api/thread/{slug_or_id}/vote`, handlers.CreateVote).Methods("POST")
	mux.HandleFunc(`/api/thread/{slug_or_id}/details`, handlers.ThreadDetails).Methods("GET")

	logHandler := logMiddleware(mux)

	log.Fatal(http.ListenAndServe(":5000", logHandler))
}