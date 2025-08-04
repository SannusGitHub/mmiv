package main

import (
	"fmt"
	"html/template"
	"log"
	"mmiv/controller"
	"net/http"
)

func main() {
	controller.OpenSQL()
	defer controller.CloseSQL()

	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	tmpl := template.Must(template.ParseFiles("./static/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token := controller.GetCookie(r, "userSessionToken")
		username := controller.GetUsernameFromCookie(r, "userSessionToken")
		id := controller.QueryFromSQL("SELECT id FROM USERS WHERE username = ?", username)
		isAdmin := controller.DoesUserMatchRank(r, "2")

		if token == nil || username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tmpl.Execute(w, map[string]interface{}{
			"Username": username,
			"Id":       id,
			"CSS":      "index.css",
			"JS":       "index.js",
			"IsAdmin":  isAdmin,
		})
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/login.html")
	})

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		controller.Login(w, r)
	})

	http.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		controller.Logout(w, r)
	})

	http.HandleFunc("/api/addPost", func(w http.ResponseWriter, r *http.Request) {
		controller.AddPost(w, r)
	})

	http.HandleFunc("/api/requestPost", func(w http.ResponseWriter, r *http.Request) {
		controller.RequestPost(w, r)
	})

	http.HandleFunc("/api/addComment", func(w http.ResponseWriter, r *http.Request) {
		controller.AddComment(w, r)
	})

	http.HandleFunc("/api/requestComment", func(w http.ResponseWriter, r *http.Request) {
		controller.RequestComment(w, r)
	})

	http.HandleFunc("/api/addUser", func(w http.ResponseWriter, r *http.Request) {
		controller.AddUser(w, r)
	})

	http.HandleFunc("/api/deleteUser", func(w http.ResponseWriter, r *http.Request) {
		controller.DeleteUser(w, r)
	})

	fmt.Println("running on http://localhost:1759/")
	if err := http.ListenAndServe("localhost:1759", nil); err != nil {
		log.Fatal(err)
	}
}
