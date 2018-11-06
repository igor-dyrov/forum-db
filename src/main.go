package main

import (
	"net/http"
	"log"
	"fmt"
	mux2 "github.com/gorilla/mux"
	"./handlers"
	"./getters"
	_ "github.com/lib/pq"
	"database/sql"
)

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s host = localhost port = 5432 sslmode=disable",
			"docker", "docker", "forum")
		var err error
		_, err = sql.Open("postgres", dbInfo)
		if err != nil {
			fmt.Printf(err.Error())
		}
		fmt.Printf("URL: %v; Method: %v; Origin: %v\n", r.URL.Path, r.Method, r.Header.Get("Origin"))
		next.ServeHTTP(w, r)
	})
}

func main() {
	getters.GetPathById(0)
	mux := mux2.NewRouter()

	mux.HandleFunc(`/api/user/{nick}/create`, handlers.CreateUser).Methods("POST")
	mux.HandleFunc(`/api/user/{nick}/profile`, handlers.GetUser).Methods("GET")
	mux.HandleFunc(`/api/user/{nick}/profile`, handlers.UpdateUser).Methods("POST")

	mux.HandleFunc(`/api/forum/create`, handlers.CreateForum).Methods("POST")
	mux.HandleFunc(`/api/forum/{slug}/details`, handlers.GetForum).Methods("GET")
	mux.HandleFunc(`/api/forum/{slug}/create`, handlers.CreateThread).Methods("POST")
	mux.HandleFunc(`/api/forum/{slug}/threads`, handlers.GetThreads).Methods("GET")
	mux.HandleFunc(`/api/forum/{slug}/users`, handlers.GetThreadUsers).Methods("GET")

	mux.HandleFunc(`/api/thread/{slug_or_id}/create`, handlers.CreatePosts).Methods("POST")
	mux.HandleFunc(`/api/thread/{slug_or_id}/vote`, handlers.CreateVote).Methods("POST")
	mux.HandleFunc(`/api/thread/{slug_or_id}/details`, handlers.ThreadDetails).Methods("GET")
	mux.HandleFunc(`/api/thread/{slug_or_id}/details`, handlers.UpdateThread).Methods("POST")

	mux.HandleFunc(`/api/thread/{slug_or_id}/posts`, handlers.GetThreadPosts).Methods("GET")

	mux.HandleFunc(`/api/post/{id}/details`, handlers.GetPost).Methods("GET")
	mux.HandleFunc(`/api/post/{id}/details`, handlers.UpdatePost).Methods("POST")

	mux.HandleFunc(`/api/service/status`, handlers.GetStatus).Methods("GET")
	mux.HandleFunc(`/api/service/clear`, handlers.ClearAll).Methods("POST")

	logHandler := logMiddleware(mux)

	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s host = localhost port = 5432 sslmode=disable",
		"docker", "docker", "forum")
	var err error
	_, err = sql.Open("postgres", dbInfo)
	if err != nil {
		log.Println(err)
	}

	log.Fatal(http.ListenAndServe(":5000", logHandler))
}