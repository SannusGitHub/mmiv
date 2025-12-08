package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var EmoticonSet map[string]bool

func LoadEmoticonsFromDB() error {
	rows, err := db.Query("SELECT name FROM emoticons")
	if err != nil {
		return err
	}
	defer rows.Close()

	EmoticonSet = make(map[string]bool)

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}

		name = strings.TrimSuffix(name, ".png")
		EmoticonSet[name] = true
	}

	return rows.Err()
}

// TODO: add a proper message for the front-end / error handling whenever emoticon can or can't be manipulated

func AddEmoticon(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		fmt.Printf("Rank mismatch in AddEmoticon, invalid perms!\n")
		http.Error(w, "Invalid permission trying to add new emoticon!", http.StatusUnsupportedMediaType)
		return
	}

	emoticonName := r.FormValue("emoticon-name")
	WriteToSQL(`
		INSERT INTO emoticons (name)
		VALUES (?)
	`, emoticonName)
	LoadEmoticonsFromDB()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func DeleteEmoticon(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		http.Error(w, "Invalid permission trying to delete emoticon!", http.StatusUnsupportedMediaType)
		return
	}

	emoticonName := r.FormValue("emoticon-name")
	WriteToSQL(`
		DELETE FROM emoticons WHERE name = ?
	`, emoticonName)
	LoadEmoticonsFromDB()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func RegexEmoticons(contentString string) string {
	var emoticonRegex = regexp.MustCompile(`:([a-zA-Z0-9_+-]+):`)
	contentString = emoticonRegex.ReplaceAllStringFunc(contentString, func(match string) string {
		name := match[1 : len(match)-1]
		if EmoticonSet[name] {
			return fmt.Sprintf(`<img class="emoticon" src="/static/img/emoticons/%s.png" alt=":%s:">`, name, name)
		}

		return match
	})

	return contentString
}
