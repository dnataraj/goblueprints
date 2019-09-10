package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

type poll struct {
	ID      *primitive.ObjectID `bson:"_id" json:"id"`
	Title   string              `json:"title"`
	Options []string            `json:"options"`
	Results map[string]int      `json:"results,omitempty"`
	APIKey  string              `json:"apikey"`
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
	case "OPTIONS":
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		respond(w, r, http.StatusOK, nil)
		return
	}
	respondHHTTPErr(w, r, http.StatusNotFound)
}

func (s *Server) handlePollsGet(w http.ResponseWriter, r *http.Request) {
	//ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	c := s.client.Database("ballots").Collection("polls")
	path := NewPath(r.URL.Path)
	var err error
	var result []*poll
	if path.HasID() {
		// get specific poll
		log.Println("fetching poll with id :", path.ID)
		id, _ := primitive.ObjectIDFromHex(path.ID)
		var p poll
		err = c.FindOne(r.Context(), bson.D{{"_id", id}}).Decode(&p)
		if err == nil {
			result = append(result, &p)
		}
	} else {
		cur, err := c.Find(r.Context(), bson.D{})
		if err == nil {
			err = cur.All(r.Context(), &result)
		}
	}
	if err != nil {
		log.Println("error extracting cursor: ", err)
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, &result)
}

func (s *Server) handlePollsPost(w http.ResponseWriter, r *http.Request) {
	//respondErr(w, r, http.StatusInternalServerError, errors.New("not implemented"))
	//ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	c := s.client.Database("ballots").Collection("polls")
	var p poll
	log.Println("creating new poll: ", p)
	if err := decodeBody(r, &p); err != nil {
		respondErr(w, r, http.StatusBadRequest, "failed to read poll from request", err)
		return
	}
	apikey, ok := APIKey(r.Context())
	if ok {
		p.APIKey = apikey
	}
	//var id primitive.ObjectID
	var id = primitive.NewObjectID()
	p.ID = &id
	p.Results = map[string]int{}

	if res, err := c.InsertOne(r.Context(), p); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "failed to insert poll", err)
		id := res.InsertedID.(primitive.ObjectID)
		log.Println("created poll with id: ", id.Hex())
		return
	}
	w.Header().Set("Location", fmt.Sprintf("polls/%s", p.ID.Hex()))
	respond(w, r, http.StatusCreated, nil)
}

func (s *Server) handlePollsDelete(w http.ResponseWriter, r *http.Request) {
	//respondErr(w, r, http.StatusInternalServerError, errors.New("not implemented"))
	//ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	c := s.client.Database("ballots").Collection("polls")
	path := NewPath(r.URL.Path)
	if !path.HasID() {
		respondErr(w, r, http.StatusInternalServerError, "cannot delete all polls")
		return
	}
	id, _ := primitive.ObjectIDFromHex(path.ID)
	log.Println("attempting to delete poll with id: ", id)
	if _, err := c.DeleteOne(r.Context(), bson.D{{"_id", id}}); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "failed to delete poll", err)
		return
	}
	respond(w, r, http.StatusOK, nil)
}
