package main

import (
	"context"
	"errors"
	"github.com/globalsign/mgo/bson"
	bson2 "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"time"
)

type poll struct {
	ID      bson.ObjectId  `bson:"_id" json:"id"`
	Title   string         `json:"title"`
	Options []string       `json:"options"`
	Results map[string]int `json:"results,omitempty"`
	APIKey  string         `json:"apikey"`
}

func (s *Server) handlePolls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handlePollsGet(w, r)
		return
	case "POST":
		s.handlePollsPost(w, r)
		return
	case "DELETE":
		s.handlePollsDelete(w, r)
		return
	}
	respondHHTTPErr(w, r, http.StatusNotFound)
}

func (s *Server) handlePollsGet(w http.ResponseWriter, r *http.Request) {
	//respondErr(w, r, http.StatusInternalServerError, errors.New("not implemented"))
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	c := s.client.Database("ballots").Collection("polls")
	path := NewPath(r.URL.Path)
	var cur *mongo.Cursor
	var err error
	var result []*poll
	if path.HasID() {
		// get specific poll
		err = c.FindOne(r.Context(), bson.D{{"_id", path.ID}}).Decode(&result)
	} else {
		cur, err = c.Find(ctx, bson2.D{})
	}
	if err != nil {
		log.Println("error extracting cursor: ", err)
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	var p poll
	for cur.Next(ctx) {
		err = cur.Decode(&p)
		log.Println("found poll: ", cur.ID())
		result = append(result, &p)
	}
	respond(w, r, http.StatusOK, &result)
}

func (s *Server) handlePollsPost(w http.ResponseWriter, r *http.Request) {
	respondErr(w, r, http.StatusInternalServerError, errors.New("not implemented"))
}

func (s *Server) handlePollsDelete(w http.ResponseWriter, r *http.Request) {
	respondErr(w, r, http.StatusInternalServerError, errors.New("not implemented"))
}
