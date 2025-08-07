package main

import (
	"fmt"
	"html/template"
	"mmiv/controller"
	"net/http"
	"os"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	controller.OpenSQL()
	defer controller.CloseSQL()

	// img directory for handling on-platform images
	mux.HandleFunc("/static/img/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/static/img/"):]
		filePath := "./static/img/" + path

		// give permission to every admin rank to view dir raw
		if controller.DoesUserMatchRank(r, "2") {
			http.ServeFile(w, r, filePath)
			return
		}

		// ...however give every user the permission to view only the files themselves
		if controller.DoesUserMatchRank(r, "1") {
			if path == "" || strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, "/404", http.StatusSeeOther)
				return
			}

			info, err := os.Stat(filePath)
			if err != nil || info.IsDir() {
				http.Redirect(w, r, "/404", http.StatusSeeOther)
				return
			}

			http.ServeFile(w, r, filePath)
			return
		}

		http.Redirect(w, r, "/404", http.StatusSeeOther)
	})

	// uploads directory for handling on-platform uploads
	mux.HandleFunc("/uploads/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/uploads/"):]
		filePath := "./uploads/" + path

		// give permission to every admin rank to view dir raw
		if controller.DoesUserMatchRank(r, "2") {
			http.ServeFile(w, r, filePath)
			return
		}

		// ...however give every user the permission to view only the files themselves
		if controller.DoesUserMatchRank(r, "1") {
			if path == "" || strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, "/404", http.StatusSeeOther)
				return
			}

			info, err := os.Stat(filePath)
			if err != nil || info.IsDir() {
				http.Redirect(w, r, "/404", http.StatusSeeOther)
				return
			}

			http.ServeFile(w, r, filePath)
			return
		}

		http.Redirect(w, r, "/404", http.StatusSeeOther)
	})

	// login directory for handling login front-end files
	mux.Handle("/static/login/", http.StripPrefix("/static/login/", http.FileServer(http.Dir("./static/login"))))

	// home directory for handling access to main website stuff
	mux.HandleFunc("/static/home/", func(w http.ResponseWriter, r *http.Request) {
		if !controller.DoesUserMatchRank(r, "1") {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		fs := http.FileServer(http.Dir("./static/home"))
		http.StripPrefix("/static/home/", fs).ServeHTTP(w, r)
	})

	// private directory for handling access to stuff the average user / stranger shouldn't prolly accewss
	mux.HandleFunc("/static/private/", func(w http.ResponseWriter, r *http.Request) {
		if !controller.DoesUserMatchRank(r, "2") {
			http.Redirect(w, r, "/404", http.StatusSeeOther)
			return
		}

		fs := http.FileServer(http.Dir("./static/private"))
		http.StripPrefix("/static/private/", fs).ServeHTTP(w, r)
	})

	// 404 directory for handling access to custom 404 page
	mux.HandleFunc("/static/404/", func(w http.ResponseWriter, r *http.Request) {
		if !controller.DoesUserMatchRank(r, "1") {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		fs := http.FileServer(http.Dir("./static/404"))
		http.StripPrefix("/static/404/", fs).ServeHTTP(w, r)
	})

	// 404 page serve function
	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		token := controller.GetCookie(r, "userSessionToken")
		username := controller.GetUsernameFromCookie(r, "userSessionToken")

		if token == nil || username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		http.ServeFile(w, r, "./static/404/404.html")
	})

	// home page serve section
	tmpl := template.Must(template.ParseFiles("./static/home/index.html"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			r.URL.Path = "/404"
			mux.ServeHTTP(w, r)
			return
		}

		token := controller.GetCookie(r, "userSessionToken")
		username := controller.GetUsernameFromCookie(r, "userSessionToken")
		id := controller.QueryFromSQL("SELECT id FROM USERS WHERE username = ?", username)
		isAdmin := controller.DoesUserMatchRank(r, "2")

		if token == nil || username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tmpl.Execute(w, map[string]any{
			"Username": username,
			"Id":       id,
			"CSS":      "index.css",
			"JS":       "index.js",
			"IsAdmin":  isAdmin,
		})
	})

	// login page serve section
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/login/login.html")
	})

	// api calls
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		controller.Login(w, r)
	})

	mux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		controller.Logout(w, r)
	})

	mux.HandleFunc("/api/addPost", func(w http.ResponseWriter, r *http.Request) {
		controller.AddPost(w, r)
	})

	mux.HandleFunc("/api/deletePost", func(w http.ResponseWriter, r *http.Request) {
		controller.DeletePost(w, r)
	})

	mux.HandleFunc("/api/requestPost", func(w http.ResponseWriter, r *http.Request) {
		controller.RequestPost(w, r)
	})

	mux.HandleFunc("/api/pinPost", func(w http.ResponseWriter, r *http.Request) {
		controller.PinPost(w, r)
	})

	mux.HandleFunc("/api/addComment", func(w http.ResponseWriter, r *http.Request) {
		controller.AddComment(w, r)
	})

	mux.HandleFunc("/api/deleteComment", func(w http.ResponseWriter, r *http.Request) {
		controller.DeleteComment(w, r)
	})

	mux.HandleFunc("/api/requestComment", func(w http.ResponseWriter, r *http.Request) {
		controller.RequestComment(w, r)
	})

	mux.HandleFunc("/api/addUser", func(w http.ResponseWriter, r *http.Request) {
		controller.AddUser(w, r)
	})

	mux.HandleFunc("/api/deleteUser", func(w http.ResponseWriter, r *http.Request) {
		controller.DeleteUser(w, r)
	})

	// run server
	fmt.Println("running on http://localhost:1759/")
	http.ListenAndServe("localhost:1759", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			r.URL.Path = "/404"
			mux.ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	}))
}
