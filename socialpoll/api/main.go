package main

import (
	"context"
	"flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

type contextKey struct {
	name string
}

var contextKeyAPIKey = &contextKey{"api-key"}

type Server struct {
	client *mongo.Client
}

func APIKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(contextKeyAPIKey).(string)
	return key, ok
}

func main() {
	var (
		addr   = flag.String("addr", ":8080", "endpoint address")
		dbaddr = flag.String("mongo", "mongodb://localhost", "mongodb address")
	)
	log.Println("dialing mongo: ", *dbaddr)
	ctx, _ := context.WithCancel(context.Background())
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(*dbaddr))
	if err != nil {
		log.Fatalln("failed to connect to mongo: ", err)
	}
	defer func() {
		log.Println("closing db connection...")
		err := client.Disconnect(ctx)
		if err != nil {
			log.Println("dialing mongodb: failed to disconnect client")
		}
		return
	}()
	s := &Server{client: client}
	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCORS(withAPIKey(s.handlePolls)))
	log.Println("starting web server on: ", *addr)
	http.ListenAndServe(":8080", mux)
	log.Println("stopping...")
}

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(w, r)
	}
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !isValidAPIKey(key) {
			respondErr(w, r, http.StatusUnauthorized, "invalid API key")
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyAPIKey, key)
		fn(w, r.WithContext(ctx))
	}
}

func isValidAPIKey(key string) bool {
	return key == "f00bar1"
}
