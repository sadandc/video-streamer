# video-streamer

## How to Run
**Install Ffmpeg**
- ✅ Windows : winget install ffmpeg
- ✅ Linux : sudo apt install ffmpeg
- ✅ Mac :brew install ffmpeg

**Running Golang**
- ```go get```
- ```go run main.go -> localhost:8080```
- ```Example route : localhost:8080/video?s=sample_1280x720_surfing_with_audio.flv```
> Sample videos available localy in **storage/videos** folder

## Caching Strategy 

✅ **Step 1 — Cache Location Check**  
- When a request comes in, hash the `location` (using sha256) to generate a unique cache key  
- Check if this transcoded video already exists in a local `storage/cache` folder.

✅ **Step 2 — Serve from Cache**  
- If the file is found in the cache folder, serve it directly with streaming to avoid FFmpeg overhead.

✅ **Step 3 — Transcode and Save to Cache**  
- If the cache file does not exist:
- Fetch the original video
- Transcode with FFmpeg (resizing to 200px height)
- Stream the result to the client immediately (*pipe* in chunks)
- Simultaneously save the transcoded video to the `storage/cache` folder for future reuse

✅ **Step 4 — CDN Integration**  
- upload the transcoded result to object storage (e.g., S3, Alibaba OSS)  
- Configure CDN like Cloudflare or else to point to that storage   
---
