package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type AnnouncementData struct {
	Content string `json:"content"`
}

/*
NOTE: currently only has support for one announcement, maybe add a history or making announcements
for later
*/
func AddAnnouncement(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		fmt.Printf("Rank mismatch in AddAnnouncement, invalid perms!\n")
		http.Error(w, "Invalid permission trying to add announcement!", http.StatusUnsupportedMediaType)
		return
	}

	content := r.FormValue("content")

	WriteToSQL(`
		INSERT INTO ANNOUNCEMENTS (content)
		VALUES (?)
	`, content)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func RemoveAnnouncement(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		http.Error(w, "Invalid permission trying to remove announcement!", http.StatusUnsupportedMediaType)
		return
	}

	WriteToSQL(`
		DELETE FROM announcements
	`)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func RequestAnnouncement(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "1") {
		fmt.Printf("Rank mismatch in RequestAnnouncement, invalid perms!\n")
		http.Error(w, "Invalid permission trying to request announcement!", http.StatusUnsupportedMediaType)
		return
	}

	var content string
	query := `SELECT content FROM announcements ORDER BY updated_at DESC LIMIT 1`
	err := db.QueryRow(query).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			content = ""
		} else {
			log.Fatal(err)
			return
		}
	}

	response := AnnouncementData{
		Content: content,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
