package controller

import (
	"bufio"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

/*
	these are just general helper functions that are too general to fit into any niche category in this project:

	stuff like checking variables, functions for comparisons that might need to be done in multiple situations,
	anything that doesn't fit into specific main post stuff
*/

type Config struct {
	ServerAddress string
	ServerPort    string
}

var Cfg *Config

func LoadConfig() {
	_ = godotenv.Load()

	Cfg = &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", "localhost"),
		ServerPort:    getEnv("SERVER_PORT", "1759"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ParseBoolOrFalse(val string) bool {
	if val == "true" {
		return true
	}
	if val == "false" {
		return false
	}
	return false
}

func GetFileLine(fileName string, lineNumber int) string {
	myfile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer myfile.Close()

	scanner := bufio.NewScanner(myfile)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == lineNumber {
			return scanner.Text()
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from file:", err)
	}

	return ""
}

func CountFileLines(fileName string) int {
	myfile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return 0
	}
	defer myfile.Close()

	scanner := bufio.NewScanner(myfile)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}

func TruncateFilename(original string, maxBaseLen int) string {
	ext := filepath.Ext(original)
	base := strings.TrimSuffix(original, ext)

	base = regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(base, "_")

	if len(base) > maxBaseLen {
		base = base[:maxBaseLen]
	}

	return base + ext
}

func IsAcceptedMIME(file multipart.File, acceptedMIMEs []string) bool {
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	file.Seek(0, io.SeekStart)

	mime := http.DetectContentType(buf[:n])

	for _, m := range acceptedMIMEs {
		if mime == m {
			return true
		}
	}
	return false
}

func IsAcceptedFileFormat(filename string, acceptedFormats []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, format := range acceptedFormats {
		if ext == format {
			return true
		}
	}

	return false
}
