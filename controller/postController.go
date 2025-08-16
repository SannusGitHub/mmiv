package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

/*
NOTE: looking back on this, maybe it would have been more useful to have more flexible code that
would allow for both "Comment" posts and "Post" posts to be unified

maybe it'll come in use eventually to have these both separated, a.la post-only or comment-only
feature but right now I highly doubt it

ALSO, there should probably be a check from the back-end to the front-end that cuts out anything the
"user" ( rank 1 ) shouldn't need vs "admin" ( rank 2 ) needs, to avoid unnecessary data being sent &
also to avoid people from snooping variables and potentially exploiting vulnerabilities because they
know what the back-end has

NOTE: also, should probably figure out how ratelimiting works in order to avoid api-spam slop
and other stuff that may degrade the quality of the platform in some way
*/

type PostData struct {
	Id           string `json:"id"`
	Username     string `json:"username"`
	PostContent  string `json:"postcontent"`
	Imagepath    string `json:"imagepath"`
	Timestamp    string `json:"timestamp"`
	CommentCount string `json:"commentcount"`
	Pinned       bool   `json:"pinned"`
	Locked       bool   `json:"locked"`
	CanPin       *bool  `json:"canpin,omitempty"`
	CanLock      *bool  `json:"canlock,omitempty"`
	HasOwnership *bool  `json:"hasownership,omitempty"`
}

func AddPost(w http.ResponseWriter, r *http.Request) {
	/*
		NOTE : random name generator for anon usernames?
		randomLine := controller.GetFileLine(
			"static/home/names.txt",
			rand.Intn(controller.CountFileLines("static/home/names.txt")),
		)
		fmt.Printf("Random line: %s\n", randomLine)
	*/

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
	isAnonymous := ParseBoolOrFalse(r.FormValue("isanonymous"))

	var locked bool
	var pinned bool
	if DoesUserMatchRank(r, "2") {
		locked = ParseBoolOrFalse(r.FormValue("locked"))
		pinned = ParseBoolOrFalse(r.FormValue("pinned"))
	} else {
		locked = false
		pinned = false
	}

	WriteToSQL(`
		INSERT INTO POSTS (id, username, postcontent, imagepath, locked, pinned, isanonymous)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, currentUsername, postContent, imagePath, locked, pinned, isAnonymous)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	var data PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	postOwner := QueryFromSQL(`SELECT username FROM posts WHERE id = ?`, data.Id)
	if !DoesUserMatchRank(r, "2") {
		if currentUsername != postOwner {
			fmt.Println("DeletePost request discarded due to invalid perms")
			return
		}
	}

	postImagePath := QueryFromSQL(`SELECT imagepath FROM posts WHERE id = ?`, data.Id)
	if postImagePath != "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
			return
		}

		joinedImagePath := filepath.Join(cwd, postImagePath)
		fmt.Println(joinedImagePath)
		err = os.Remove(joinedImagePath)

		if err != nil {
			fmt.Printf("File of %s was not found...\n", joinedImagePath)
		}
	}

	WriteToSQL(`DELETE FROM posts WHERE id = ?`, data.Id)
	fmt.Printf("Post ID %s deleted successfully\n", data.Id)

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

	query := `SELECT id, username, postcontent, imagepath, timestamp, pinned, locked, isanonymous FROM POSTS ORDER BY pinned DESC, id DESC`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var posts []PostData

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	for rows.Next() {
		var post PostData
		var isAnonymous bool

		err := rows.Scan(
			&post.Id,
			&post.Username,
			&post.PostContent,
			&post.Imagepath,
			&post.Timestamp,
			&post.Pinned,
			&post.Locked,
			&isAnonymous,
		)
		if err != nil {
			log.Fatal(err)
		}

		countQuery := `SELECT COUNT(*) FROM comments WHERE parentpostid = ?`
		err = db.QueryRow(countQuery, post.Id).Scan(&post.CommentCount)
		if err != nil {
			log.Printf("Error counting comments for post ID %s: %v\n", post.Id, err)
			post.CommentCount = "0"
		}

		// hidden name case
		if isAnonymous {
			if DoesUserMatchRank(r, "2") {
				post.Username = post.Username + " (hidden)"
			} else {
				post.Username = "hidden"
			}
		}

		var hasOwnership bool
		if currentUsername == post.Username || DoesUserMatchRank(r, "2") {
			hasOwnership = true
			post.HasOwnership = &hasOwnership
			// fmt.Printf("Post of ID %s is owned by requester\n", post.Id)
		}

		var canPin bool
		var canLock bool
		if DoesUserMatchRank(r, "2") {
			canPin = true
			canLock = true
			post.CanPin = &canPin
			post.CanLock = &canLock
		}

		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func PinPost(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		fmt.Printf("Rank mismatch in PinPost, invalid perms!\n")
		return
	}

	var data PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	WriteToSQL(`UPDATE posts SET pinned = ? WHERE id = ?`, data.Pinned, data.Id)
	fmt.Printf("Post of ID %s has been pinned: %t\n", data.Id, data.Pinned)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func LockPost(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		fmt.Printf("Rank mismatch in LockPost, invalid perms!\n")
		return
	}

	var data PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	WriteToSQL(`UPDATE posts SET locked = ? WHERE id = ?`, data.Locked, data.Id)

	fmt.Printf("Post of ID %s has been locked: %t\n", data.Id, data.Locked)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

type CommentData struct {
	Id           string `json:"id"`
	ParentPostID string `json:"parentpostid"`
	Username     string `json:"username"`
	PostContent  string `json:"postcontent"`
	Imagepath    string `json:"imagepath"`
	Timestamp    string `json:"timestamp"`
	IsComment    bool   `json:"iscomment"`
	HasOwnership *bool  `json:"hasownership,omitempty"`
}

func AddComment(w http.ResponseWriter, r *http.Request) {
	/*
		NOTE: it would probably be a smart idea to add an if-check for whether
		parentpostid actually constitutes as a post or not
	*/
	if !DoesUserMatchRank(r, "1") {
		fmt.Printf("Rank mismatch in AddComment, invalid perms!\n")
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Fatal(err)
		return
	}

	// untested, shooould check for if parentpostid is a valid id in posts?
	parentID, err := strconv.Atoi(r.FormValue("parentpostid"))
	if err != nil {
		http.Error(w, "Invalid parent post ID", http.StatusBadRequest)
		return
	}

	var isLocked bool
	err = db.QueryRow(`SELECT locked FROM posts WHERE id = ?`, parentID).Scan(&isLocked)
	if err != nil {
		log.Printf("Database error checking parent post: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if isLocked && !DoesUserMatchRank(r, "2") {
		fmt.Println("Post is locked, not replying...")
		return
	}

	var exists bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)`, parentID).Scan(&exists)
	if err != nil {
		log.Printf("Database error checking parent post: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Parent post does not exist", http.StatusBadRequest)
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
	isAnonymous := ParseBoolOrFalse(r.FormValue("isanonymous"))
	ParentPostID := r.FormValue("parentpostid")

	WriteToSQL(`
		INSERT INTO COMMENTS (id, parentpostid, username, postcontent, imagepath, isanonymous)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, ParentPostID, currentUsername, postContent, imagePath, isAnonymous)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// untested
func DeleteComment(w http.ResponseWriter, r *http.Request) {
	var data CommentData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	postOwner := QueryFromSQL(`SELECT username FROM posts WHERE id = ?`, data.Id)
	if !DoesUserMatchRank(r, "2") {
		if currentUsername != postOwner {
			fmt.Println("DeleteComment request discarded due to invalid perms")
			return
		}
	}

	commentImagePath := QueryFromSQL(`SELECT imagepath FROM comments WHERE id = ?`, data.Id)
	if commentImagePath != "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
			return
		}

		joinedImagePath := filepath.Join(cwd, commentImagePath)
		fmt.Println(joinedImagePath)
		err = os.Remove(joinedImagePath)

		if err != nil {
			log.Fatal(err)
		}
	}

	WriteToSQL(`DELETE FROM comments WHERE id = ?`, data.Id)
	fmt.Printf("Comment ID %s deleted successfully\n", data.Id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func RequestComment(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "1") {
		fmt.Printf("Rank mismatch in RequestComments, invalid perms!\n")
		return
	}

	var data CommentData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	query := `SELECT id, username, postcontent, imagepath, timestamp, isanonymous FROM COMMENTS WHERE parentpostid = ?`
	rows, err := db.Query(query, data.ParentPostID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var comments []CommentData

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	for rows.Next() {
		var comment CommentData
		var isAnonymous bool

		err := rows.Scan(
			&comment.Id,
			&comment.Username,
			&comment.PostContent,
			&comment.Imagepath,
			&comment.Timestamp,
			&isAnonymous,
		)
		if err != nil {
			log.Fatal(err)
		}

		if isAnonymous {
			if DoesUserMatchRank(r, "2") {
				comment.Username = comment.Username + " (hidden)"
			} else {
				comment.Username = "hidden"
			}
		}

		var hasOwnership bool
		if currentUsername == comment.Username || DoesUserMatchRank(r, "2") {
			hasOwnership = true
			comment.HasOwnership = &hasOwnership
		}

		// add a marker to differentiate front-end whether a post element is a comment
		// yes i'm doing it this way. there's probably a better way. sue me.
		comment.IsComment = true

		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
