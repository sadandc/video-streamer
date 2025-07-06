package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type FuncError struct {
	Message string
	Code    int
}

func HashFileName(source string) string {
	hash := sha256.New()
	hash.Write([]byte(source))
	hashInBytes := hash.Sum(nil)

	return hex.EncodeToString(hashInBytes)
}

func FileIsExists(filePath string) bool {
	_, err := os.Stat(filePath)

	return err == nil
}

// func GetSourceReader(source string) (string) {
func GetSourceReader(source string) (io.Reader, error, string) {
	if isExternalUrl(source) {
		extResponse, err := http.Get(source)
		if err != nil {
			return nil, err, "pipe:0"
		}

		return extResponse.Body, nil, "pipe:0"
		// return source
	}

	storagePath := "./storage/videos/" + source
	videoFile, err := os.Open(storagePath)
	if err != nil {
		fmt.Println("Failed to open file")

		return nil, err, "storage"
	}
	defer videoFile.Close()

	return videoFile, nil, storagePath
}

func isExternalUrl(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func GetSourceFromUrl(request *http.Request) (string, *FuncError) {
	param := request.URL.Query().Get("s")

	if param == "" {
		return "", &FuncError{Message: "blabla", Code: http.StatusBadRequest}
	}

	return param, nil
}

func (e *FuncError) Error() string {
	return fmt.Sprintf("%s (code : %d)", e.Message, e.Code)
}
