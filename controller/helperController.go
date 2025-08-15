package controller

import (
	"bufio"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

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
