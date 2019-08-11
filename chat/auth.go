package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"log"
	"net/http"
	"strings"
)

type authHandler struct {
	next http.Handler
}

type User struct {
	Login     string
	Name      string
	HTMLURL   string `json:"html_url"`
	AvatarURL string `json:"avatar_url"`
}

// OAuth 2.0 init and config
var conf = &oauth2.Config{
	ClientID:     "b9d234ae03ae0a688e64",
	ClientSecret: "30548529437ad927b7e088ae67441555f658ac31",
	Endpoint:     github.Endpoint,
	RedirectURL:  "http://localhost:8080/auth/callback/github",
	Scopes:       nil,
}

var ctx = context.Background()

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth")
	if err == http.ErrNoCookie || cookie.Value == "" {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// success - call the next handler
	h.next.ServeHTTP(w, r)
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// Handle 3rd party login process
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		log.Println("TODO handle login for", provider)
		loginUrl := conf.AuthCodeURL("", oauth2.AccessTypeOffline)
		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "authorization code not found", http.StatusBadRequest)
			return
		}
		token, err := conf.Exchange(ctx, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		client := conf.Client(ctx, token)
		response, err := client.Get("https://api.github.com/user")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()
		//fmt.Fprintf(w, "User Info : %s\n", userInfo)
		var user User
		if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "auth", Value: base64.URLEncoding.EncodeToString(data), Path: "/"})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "auth action %s not supported", action)
	}
}

func unwrapCookie(cookie *http.Cookie) (*User, error) {
	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("retrieving auth cookie value failed: %s", err)
	}
	var user User
	if err := json.Unmarshal(value, &user); err != nil {
		return nil, fmt.Errorf("JSON unmarshalling failed: %s", err)
	}

	return &user, nil
}
