package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

/*
	anything related to posting stuff is here: creating / removing / retrieving posts alongside comments
	and their relevant data.

	NOTE: looking back on this, maybe it would have been more useful to have more flexible code that
	would allow for both "Comment" posts and "Post" posts to be unified. too bad, i've already decided
	on two specific formats

	maybe it'll come in use eventually to have these both separated, a.la post-only or comment-only
	feature but right now I highly doubt it

	ALSO, there should probably be a check from the back-end to the front-end that cuts out anything the
	"user" ( rank 1 ) shouldn't need vs "admin" ( rank 2 ) needs, to avoid unnecessary data being sent &
	also to avoid people from snooping variables and potentially exploiting vulnerabilities because they
	know what the back-end has

	NOTE: also, should probably figure out how ratelimiting works in order to avoid api-spam slop
	and other stuff that may degrade the quality of the platform in some way

	TODO:
		* page system for posts (start from nr. / cont. from nr.)
		* thumbnail compression for posts (?)
		* "message too long, click here to expand" feature for long posts
		* message length limit (250 characters maybe) (?)
		* hide post / comment

	FIXME / BUGS:
		* deletion of posts causes a display desync for some reason
		reproduction steps:
			* create a post of A
			* delete a post of A
			* create a post of B
			* appears post A
			* create a post of C
			* appears post B...

		caused by pagination with desync: offset specifically

		clicking on a post with a link both opens the new tab and also the post itself for comments

		link support for unsanitized post (regex messes it up)

*/

var acceptedExts = []string{
	".jpg", ".jpeg", ".png", ".gif",
}

var acceptedMIMEs = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
}

/*
struct for post-related data that we can assemble and serve
  - ID: ID of the post in the database, front-end displayed numerically as well
  - Username: Username of the person that created the post
  - PostContent: text string accompanied by the post
  - Imagepath: local machine path to the image that's being stored
  - Timestamp: timestamp of when the post was submitted
  - CommentCount: how many children comments the post has
  - Pinned: whether post is pinned by someone with escalated privileges (shows up top)
  - Locked: whether post is uncommentable by someone with escalated privileges (shows up top)
  - CanPin: back-end variable for when administrators are querying a post and should have the option of pinning available
  - CanLock: back-end variable for when administrators are querying a post and should have the option of locking available
  - HasOwnership: back-end variable for when a person has "ownership" of a post
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

type PostRequest struct {
	DisplayFromPostNumber int `json:"displayfrompostnumber"`
	AmountOfPostsRequired int `json:"amountofpostsrequired"`
}

/*
adding posts, AddPost() function does the following outlined:

  - checks whether the user is properly authenticated as a valid "account" member

    if not above rank "1", return invalid permission and no permission error. they're
    not allowed to post because they don't have a valid account they're logged into

  - parse the form given for adding a post and acquire the variables, this includes
    retrieving the data for and assembling the PostData struct as follows:

    userSessionToken - for usernames and anything to do with account specific details, so
    we know that it's "the person" that is trying to upload it

    postContent - string of what message the user wants to upload

    imagePath - the asset, raw image itself and where it will be uploaded to on the server
    (in this case the /uploads/ directory)

    locked, pinned, isAnonymous - self explanitory, any secondary options that the user chooses

  - get the raw image from the form (imagePath, file, handler, err under r.FormFile("image"))

  - check whether it is an accepted file format (assigned with "acceptedFileFormats" variable)

    if file does not constitute as a "valid" format from a predetermined list we return false,
    cancel adding a post and display an error message for the front-end letting them know

  - upload the file with an unique timestamp on the machine to /uploads directory

    timestamp is added to avoid any duplicate names and to trace when it was uploaded

  - check for any secondary variables (if the post has been requested to be locked, pinned, anonymized)

    and of course, default any variables to a default state whenever a person that should have
    no permissions attempts to do administrator actions like locking or pinning

  - finally run WriteToSQL(), feeding in id, currentUsername, postContent, imagePath, locked, pinned, isAnonymous

    ...and then return success to the front-end
*/
func AddPost(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "1") {
		fmt.Printf("Rank mismatch in AddPost, invalid perms!\n")
		http.Error(w, "No permission to upload post!", http.StatusForbidden)
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

		if !IsAcceptedFileFormat(handler.Filename, acceptedExts) || !IsAcceptedMIME(file, acceptedMIMEs) {
			http.Error(w, "Unsupported file format!", http.StatusUnsupportedMediaType)
			return
		}

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

	// check whether we should sanitize or not (useful if moderators want better control with injecting HTML)
	var shouldBeDesanitized = ParseBoolOrFalse(r.FormValue("reject-sanitize"))
	if !(DoesUserMatchRank(r, "2") && shouldBeDesanitized) {
		fmt.Println("Sanitizing post content to avoid malicious actions...")
		postContent = html.EscapeString(postContent)
	}

	// check for locking and pinning, whether user has auth to do it and default to false if not
	var locked, pinned bool
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

/*
refer to comment about addPost() function, generally the same vibe except reversed

  - get username of the person asking for a delete request and the ID of the post

  - check if post is either owned by the user or delete is requested by administrator,
    if neither are valid then return due to invalid permissions

  - get the image path of the post and delete it, just to avoid unnecessary storage of
    files we no longer want

  - delete from database using the ID we acquired from the request
*/
func DeletePost(w http.ResponseWriter, r *http.Request) {
	var data PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	postOwner, err := QueryFromSQL(`SELECT username FROM posts WHERE id = ?`, data.Id)
	if err != nil {
		fmt.Println("Warning: Can't get post owner! Invalidating delete...")
		return
	}

	if !DoesUserMatchRank(r, "2") {
		if currentUsername != postOwner {
			fmt.Println("DeletePost request discarded due to invalid perms")
			return
		}
	}

	postImagePath, err := QueryFromSQL(`SELECT imagepath FROM posts WHERE id = ?`, data.Id)
	if err != nil {
		fmt.Println("Warning: Can't get getting postImagePath! Defaulting to empty...", err)
		postImagePath = ""
	}

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

	// handle the first request: we ask from what index and how many posts
	displayFromStr := r.FormValue("displayfrompostnumber")
	amountReqStr := r.FormValue("amountofpostsrequested")

	displayFromInt, err := strconv.Atoi(displayFromStr)
	if err != nil {
		displayFromInt = 1
	}

	amountReqInt, err := strconv.Atoi(amountReqStr)
	if err != nil {
		amountReqInt = 20
	}

	// then we get the actual posts themselves
	query := `
		SELECT id, username, postcontent, imagepath, timestamp, pinned, locked, isanonymous
		FROM POSTS
		ORDER BY pinned DESC, id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(query, amountReqInt, displayFromInt)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No posts found, returning null...")
		} else {
			log.Fatal(err)
			return
		}
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
				post.Username = "Hidden"
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

		// deal with emoticons
		post.PostContent = RegexEmoticons(post.PostContent)

		// link support
		post.PostContent = RegexLink(post.PostContent)

		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func PinPost(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		fmt.Printf("Rank mismatch in PinPost, invalid perms!\n")
		http.Error(w, "No permission to pin post!", http.StatusForbidden)
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
		http.Error(w, "No permission to lock post!", http.StatusForbidden)
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
		http.Error(w, "No permission to upload comment!", http.StatusForbidden)
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

	// a check for whether the post is locked: return error if it is and user isn't admin
	var isLocked bool
	err = db.QueryRow(`SELECT locked FROM posts WHERE id = ?`, parentID).Scan(&isLocked)
	if err != nil {
		log.Printf("Database error checking parent post: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if isLocked && !DoesUserMatchRank(r, "2") {
		fmt.Println("Post is locked, not replying...")
		http.Error(w, "Cannot comment under post, locked!", http.StatusForbidden)
		return
	}

	// a check for whether the post exists: return error if it doesn't exist
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

		if !IsAcceptedFileFormat(handler.Filename, acceptedExts) || !IsAcceptedMIME(file, acceptedMIMEs) {
			http.Error(w, "Unsupported file format!", http.StatusUnsupportedMediaType)
			return
		}

		safeName := TruncateFilename(handler.Filename, 64)
		uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
		imagePath = filepath.Join("uploads", uniqueName)
		dst, err := os.Create(imagePath)
		if err != nil {
			fmt.Printf("Error in os.Create() of imagefile!\n")
			http.Error(w, "Encountered error with file on back-end.", http.StatusForbidden)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			fmt.Printf("Error in io.Copy() of imagefile!\n")
			http.Error(w, "Encountered error with file on back-end.", http.StatusForbidden)
			return
		}
	} else if err != http.ErrMissingFile {
		fmt.Printf("Error, image file seems to be missing!\n")
		http.Error(w, "Encountered error with file on back-end.", http.StatusForbidden)
		return
	}

	res, _ := db.Exec(`INSERT INTO global_ids DEFAULT VALUES`)
	id, _ := res.LastInsertId()

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	postContent := r.FormValue("postcontent")
	isAnonymous := ParseBoolOrFalse(r.FormValue("isanonymous"))
	ParentPostID := r.FormValue("parentpostid")

	// check whether we should sanitize or not (useful if moderators want better control with injecting HTML)
	var shouldBeDesanitized = ParseBoolOrFalse(r.FormValue("reject-sanitize"))
	fmt.Println("Sanitization:", shouldBeDesanitized)
	if !(DoesUserMatchRank(r, "2") && shouldBeDesanitized) {
		fmt.Println("Sanitizing comment content to avoid malicious actions...")
		postContent = html.EscapeString(postContent)
	}

	WriteToSQL(`
		INSERT INTO COMMENTS (id, parentpostid, username, postcontent, imagepath, isanonymous)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, ParentPostID, currentUsername, postContent, imagePath, isAnonymous)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	var data CommentData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	currentUsername := GetUsernameFromCookie(r, "userSessionToken")
	commentOwner, err := QueryFromSQL(`SELECT username FROM comments WHERE id = ?`, data.Id)
	if err != nil {
		fmt.Println("Warning: Can't get comment owner for some reason! commentOwner is: ", commentOwner)
		return
	}

	if !DoesUserMatchRank(r, "2") {
		if currentUsername != commentOwner {
			fmt.Println("DeleteComment request discarded due to invalid perms")
			return
		}
	}

	commentImagePath, err := QueryFromSQL(`SELECT imagepath FROM comments WHERE id = ?`, data.Id)
	if err != nil {
		fmt.Println("Warning: Can't commentImagePath! Defaulting to empty...")
		commentImagePath = ""
	}

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
		http.Error(w, "Invalid permission when trying to request comment!", http.StatusUnsupportedMediaType)
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

		// deal with emoticons
		comment.PostContent = RegexEmoticons(comment.PostContent)

		// link support
		comment.PostContent = RegexLink(comment.PostContent)

		// add a marker to differentiate front-end whether a post element is a comment
		// yes i'm doing it this way. there's probably a better way. sue me.
		comment.IsComment = true
		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func RegexLink(contentString string) string {
	linkRegex := regexp.MustCompile(`(https||http)?://[^\s]+`)

	return linkRegex.ReplaceAllStringFunc(contentString, func(url string) string {
		return fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, url, url)
	})
}
