package service

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"go-video-streamer/utils"
)

var lock sync.Map

func Ffmpeg(source string, cache string, response http.ResponseWriter) {
	// buat file lock agar tidak conflict jika ada request yang emngakses video yang sama
	lockError := continueOrLock(cache)
	if lockError != nil {
		log.Println("ffmpeg exited with error:", lockError.Message)

		http.Error(response, lockError.Message, lockError.Code)

		return
	}

	// Dapatkan file Reader dari file local atau URL
	getSource, err, sourceName := utils.GetSourceReader(source)
	// getSource := utils.GetSourceReader(source)
	if err != nil {
		log.Println("ffmpeg exited with error:", err)

		http.Error(response, err.Error(), http.StatusInternalServerError)

		return
	}

	if flusher, ok := response.(http.Flusher); ok {
		// optional: flush dulu biar cepat respon
		flusher.Flush()
	}

	// Jalankan ffmpeg menggunakan pipe agar bisa di stream langsung ke client
	ffmpeg := exec.Command("ffmpeg",
		"-i", sourceName,
		"-vf", "scale=-1:200",
		"-c:v", "libx264",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof",
		"-f", "mp4",
		"-preset", "fast",
		"pipe:1",
	)

	// persiapkan stdin
	stdin, err := ffmpeg.StdinPipe()
	if err != nil {
		log.Println("ffmpeg error : Stdin error", err)

		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	// persiapkan stdout
	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Println("Failed to pipe")

		http.Error(response, "ffmpeg error : stdout error", http.StatusInternalServerError)
		return
	}

	//persiapkan stderr untuk log
	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		log.Println("Failed to pipe")

		http.Error(response, "ffmpeg error : stderr error", http.StatusInternalServerError)

		return
	}

	// jalankan ffmpeg sekarang
	if err := ffmpeg.Start(); err != nil {
		log.Println("Failed to pipe")

		http.Error(response, "ffmpeg error : stdout error", http.StatusInternalServerError)

		return
	}

	// Baca file sumber dengan di chunk lalu proses dengan ffmpeg
	go func() {
		_, err = io.Copy(stdin, getSource)
		stdin.Close()
	}()

	// Buat log dari ffmpeg
	go func() {
		io.Copy(log.Writer(), stderr)
	}()

	// buat dulu file Cache di storage/cache
	cacheFile, err := os.Create(cache)
	if err != nil {
		log.Println("Failed to Start Ffmpeg")

		http.Error(response, "ffmpeg error : Failed to start", http.StatusInternalServerError)

		return
	}
	defer cacheFile.Close()

	// Baca hasil dari ffmpeg yang sudah di proses, lalu simpan ke file cache
	tee := io.TeeReader(stdout, cacheFile)

	// Copy Hasil ffmpeg untuk di stream ke response (client)
	if _, err := io.Copy(response, tee); err != nil {
		log.Println("Failed to create output to client")

		http.Error(response, "ffmpeg error : faled to stream", http.StatusInternalServerError)

		return
	}

	// Tunggu ffmpeg sampai selesai memproses semuanya
	if err := ffmpeg.Wait(); err != nil {
		log.Println("ffmpeg exited with error:", err)

		http.Error(response, "ffmpeg error : exited with error"+err.Error(), http.StatusInternalServerError)

		return
	}
}

// Check Cache file
func ValidCacheFile(source string) (string, bool) {
	fileName := "./storage/cache/" + utils.HashFileName(source) + ".mp4"

	return fileName, utils.FileIsExists(fileName)
}

// Buat file lock
func continueOrLock(cacheFileName string) *utils.FuncError {
	lockFile := cacheFileName + ".lock"

	_, isExists := lock.LoadOrStore(lockFile, true)
	if isExists {
		return &utils.FuncError{Message: "Server is busy, please retry later", Code: http.StatusTooEarly}
	}

	defer lock.Delete(lockFile)

	return nil
}

// Baca File yang sudah pernah di buat dari Ffmpeg, lalu stream ke client dengan cara di chunk
func StreamCacheFile(cacheFile string, response http.ResponseWriter) *utils.FuncError {
	cf, errCache := os.Open(cacheFile)
	if errCache != nil {
		return &utils.FuncError{Message: "Cache File Not Found", Code: http.StatusNotFound}
	}
	defer cf.Close()

	if flusher, ok := response.(http.Flusher); ok {
		// optional: flush dulu biar cepat respon
		flusher.Flush()
	}

	_, errCache = io.Copy(response, cf)
	if errCache != nil {
		return &utils.FuncError{Message: errCache.Error(), Code: http.StatusInternalServerError}
	}

	return nil
}
