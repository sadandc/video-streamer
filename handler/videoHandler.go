package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"go-video-streamer/service"
	"go-video-streamer/utils"

	"golang.org/x/sync/semaphore"
)

// batasi misalkan 10 ffmpeg proses
var ffmpegLimiter = semaphore.NewWeighted(10)

func Index(response http.ResponseWriter, request *http.Request) {
	http.ServeFile(response, request, "./view/index.html")
}

func VideoHandler(response http.ResponseWriter, request *http.Request) {
	// Set Header agar pada response Content-length nya bisa menyesuaikan dengan chunk
	response.Header().Set("Content-Type", "video/mp4")
	response.Header().Set("Transfer-Encoding", "chunked")
	response.WriteHeader(http.StatusOK)

	// Dapatkan url dari query string parameter
	// variable nya adalah "s", e.g : /video?s=url
	source, err := utils.GetSourceFromUrl(request)
	if err != nil {
		log.Println("Get source url error :", err)
		http.Error(response, err.Message, err.Code)
	}

	// cek apakah file yang sudah di resize oleh ffmpeg ada atau tidak
	// jika ada maka tidak perlu panggil ffmpeg lagi, tinggal stream dari file yang ada
	cacheFileName, isExist := service.ValidCacheFile(source)
	if isExist {
		err = service.StreamCacheFile(cacheFileName, response)
		if err != nil {
			log.Println("Cache is not valid :", err)
			http.Error(response, err.Message, err.Code)
		}

		return
	}

	// sementara ini bisa di batasin dulu agar ffmpeg tidak terlalu banyak jika ada user yang request bersamaan
	err = limiter(request)
	if err != nil {
		log.Println("Ffmpeg is busy :", err)
		http.Error(response, err.Message, err.Code)

		return
	}

	// jalankan ffmpeg untuk resize video
	// lalu di stream ke client
	// lalu simpan filenya di storage/cache
	service.Ffmpeg(source, cacheFileName, response)
}

func limiter(request *http.Request) *utils.FuncError {
	ctx := request.Context()

	acquireCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := ffmpegLimiter.Acquire(acquireCtx, 1); err != nil {
		return &utils.FuncError{Message: "Server busy, try again later", Code: http.StatusTooManyRequests}
	}
	defer ffmpegLimiter.Release(1)

	return nil
}
