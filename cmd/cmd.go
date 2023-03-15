package cmd

import (
	"log"
	"net/http"
	"text/template"

	"github.com/yunling101/chatgpt-web/config"
	"github.com/yunling101/chatgpt-web/pkg/route"
	"github.com/yunling101/chatgpt-web/public"
)

func Listener() {
	tmpl := template.Must(template.ParseFS(public.FS, "build/*.html"))
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(public.StaticFS("/static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index.html", nil)
	})
	http.HandleFunc("/chat", route.HandleWebSocket)
	log.Printf("listen to %s", config.Listen)

	log.Fatal(http.ListenAndServe(config.Listen, nil))
}
