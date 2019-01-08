package main

import (
	"fmt"
	"log"
	"time"

	"net/http"

	mux "github.com/gorilla/mux"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/handlers"
)

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		dt := time.Since(start)

		if r.Method == "GET" {
			fmt.Printf("% 8d -> %v\n", int64(dt/time.Microsecond), r.URL.Path)
		}
	})
}

func main() {

	// runtime.GOMAXPROCS(runtime.NumCPU())

	common.GetDB()
	common.InitConnectionPool()

	router := mux.NewRouter()

	router.HandleFunc(`/api/user/{nick}/create`, handlers.CreateUser).Methods("POST")
	router.HandleFunc(`/api/user/{nick}/profile`, handlers.GetUser).Methods("GET")
	router.HandleFunc(`/api/user/{nick}/profile`, handlers.UpdateUser).Methods("POST")

	router.HandleFunc(`/api/forum/create`, handlers.CreateForum).Methods("POST")
	router.HandleFunc(`/api/forum/{slug}/details`, handlers.GetForum).Methods("GET")
	router.HandleFunc(`/api/forum/{slug}/create`, handlers.CreateThread).Methods("POST")
	router.HandleFunc(`/api/forum/{slug}/threads`, handlers.GetThreads).Methods("GET")
	router.HandleFunc(`/api/forum/{slug}/users`, handlers.GetForumUsers).Methods("GET")

	router.HandleFunc(`/api/thread/{slug_or_id}/create`, handlers.CreatePosts).Methods("POST")
	router.HandleFunc(`/api/thread/{slug_or_id}/vote`, handlers.CreateVote).Methods("POST")
	router.HandleFunc(`/api/thread/{slug_or_id}/details`, handlers.ThreadDetails).Methods("GET")
	router.HandleFunc(`/api/thread/{slug_or_id}/details`, handlers.UpdateThread).Methods("POST")

	router.HandleFunc(`/api/thread/{slug_or_id}/posts`, handlers.GetThreadPosts).Methods("GET")

	router.HandleFunc(`/api/post/{id}/details`, handlers.GetPost).Methods("GET")
	router.HandleFunc(`/api/post/{id}/details`, handlers.UpdatePost).Methods("POST")

	router.HandleFunc(`/api/service/status`, handlers.GetStatus).Methods("GET")
	router.HandleFunc(`/api/service/clear`, handlers.ClearAll).Methods("POST")

	logHandler := logMiddleware(router)

	log.Print("<Start server>")

	log.Fatal(http.ListenAndServe(":5000", logHandler))
}
