package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func main() {
	var addr = flag.String("addr", ":8080", "The address of the application.")
	flag.Parse()
	r := newRoom(UseAuthAvatar)
	//r.tracer = trace.New(os.Stdout)
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	go r.run()
	// start the web server
	log.Println("Starting the web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("listen and serve: ", err)
	}

}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		// Unmarshal data
		if user, err := unwrapCookie(authCookie); err == nil {
			data["UserData"] = *user
		}
	}
	log.Printf("data : %s\n", data)
	t.templ.Execute(w, data)
}
