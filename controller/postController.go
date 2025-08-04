package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type PostData struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	PostContent string `json:"postcontent"`
	Imagepath   string `json:"imagepath"`
	Timestamp   string `json:"timestamp"`
}

func AddPost(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "1") {
		fmt.Printf("Rank mismatch in AddPost, invalid perms!\n")
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Fatal(err)
		return
	}

	var imagePath string
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), handler.Filename)
		imagePath = filepath.Join("uploads", uniqueName)
		dst, err := os.Create(imagePath)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else if err != http.ErrMissingFile {
		log.Fatal(err)
		return
	}

	res, _ := db.Exec(`INSERT INTO global_ids DEFAULT VALUES`)
	id, _ := res.LastInsertId()

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	postContent := r.FormValue("postcontent")

	WriteToSQL(`
		INSERT INTO POSTS (id, username, postcontent, imagepath)
		VALUES (?, ?, ?, ?)
	`, id, currentUsername, postContent, imagePath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func RequestPost(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "1") {
		fmt.Printf("Rank mismatch in RequestPost, invalid perms!\n")
		return
	}

	query := `SELECT id, username, postcontent, imagepath, timestamp FROM POSTS ORDER BY id DESC`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var posts []PostData

	for rows.Next() {
		var post PostData
		err := rows.Scan(&post.Id, &post.Username, &post.PostContent, &post.Imagepath, &post.Timestamp)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

type CommentData struct {
	Id           string `json:"id"`
	ParentPostID string `json:"parentpostid"`
	Username     string `json:"username"`
	PostContent  string `json:"postcontent"`
	Imagepath    string `json:"imagepath"`
	Timestamp    string `json:"timestamp"`
}

func AddComment(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "1") {
		fmt.Printf("Rank mismatch in AddPost, invalid perms!\n")
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Fatal(err)
		return
	}

	var imagePath string
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), handler.Filename)
		imagePath = filepath.Join("uploads", uniqueName)
		dst, err := os.Create(imagePath)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else if err != http.ErrMissingFile {
		log.Fatal(err)
		return
	}

	res, _ := db.Exec(`INSERT INTO global_ids DEFAULT VALUES`)
	id, _ := res.LastInsertId()

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	postContent := r.FormValue("postcontent")
	ParentPostID := r.FormValue("parentpostid")

	WriteToSQL(`
		INSERT INTO COMMENTS (id, parentpostid, username, postcontent, imagepath)
		VALUES (?, ?, ?, ?, ?)
	`, id, ParentPostID, currentUsername, postContent, imagePath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func RequestComment(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "1") {
		fmt.Printf("Rank mismatch in RequestPost, invalid perms!\n")
		return
	}

	var data CommentData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	query := `SELECT id, username, postcontent, imagepath, timestamp FROM COMMENTS WHERE parentpostid = ?`
	rows, err := db.Query(query, data.ParentPostID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var comments []CommentData

	for rows.Next() {
		var comment CommentData
		err := rows.Scan(&comment.Id, &comment.Username, &comment.PostContent, &comment.Imagepath, &comment.Timestamp)
		if err != nil {
			log.Fatal(err)
		}
		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
