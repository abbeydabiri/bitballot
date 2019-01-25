package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/mobile/asset"
)

//Asset ...
func Asset(filename string) (assetByte []byte, assetError error) {
	if strings.HasSuffix(filename, "/") {
		assetError = fmt.Errorf("directory listing forbidden")
	} else {

		switch Get().OS {
		case "ios", "android":
			switch {
			case
				filename == "public.pem",
				filename == "private.pem",
				strings.HasPrefix(filename, "frontend/"):
				if f, errOpen := asset.Open(filename); errOpen == nil {
					defer f.Close()
					assetByte, assetError = ioutil.ReadAll(f)
				}

			default:
				assetByte, assetError = ioutil.ReadFile(Get().Path + filename)
			}
		default:
			assetByte, assetError = ioutil.ReadFile(Get().Path + filename)
		}
	}
	return
}

//AssetDir ...
func AssetDir(fileDir string) (assetString []string, assetError error) {
	var filePath string
	switch Get().OS {
	case "ios", "android":
		filePath = Get().Path + fileDir
	default:
		filePath = "." + fileDir
	}
	fileInfos, err := ioutil.ReadDir(filePath)
	assetString = make([]string, len(fileInfos))
	for counter, file := range fileInfos {
		assetString[counter] = file.Name()
	}
	assetError = err
	return
}

//AssetRemove ...
func AssetRemove(filePath string) (assetError error) {
	filePath = Get().Path + filePath
	if err := os.Remove(filePath); err != nil {
		log.Printf("AssetRemove: %v", err.Error())
	}
	return
}

//AssetDirList ...
func AssetDirList(fileDir string) (assetString []string, assetError error) {
	fileDir = Get().Path + fileDir
	fileInfos, err := ioutil.ReadDir(fileDir)
	assetString = make([]string, len(fileInfos))
	for counter, file := range fileInfos {
		assetString[counter] = file.Name()
	}
	assetError = err
	return
}

//AssetDirRemove ...
func AssetDirRemove(fileDir string) (assetError error) {
	fileDir = Get().Path + fileDir
	_, assetError = os.Stat(fileDir)
	if assetError == nil {
		os.RemoveAll(fileDir)
		if err := os.Remove(fileDir); err != nil {
			log.Printf("AssetDirRemove: %v", err.Error())
		}
	}
	return
}

//ContentType ...
func ContentType(filename string) (contentType string) {
	// contentType = "text/plain; charset=utf-8"
	contentType = "text/html"
	switch {
	case strings.HasSuffix(filename, ".apk"):
		contentType = "application/vnd.android.package-archive"

	case strings.HasSuffix(filename, ".js"):
		contentType = "application/javascript"
	case strings.HasSuffix(filename, ".json"):
		contentType = "application/json"
	case strings.HasSuffix(filename, ".pdf"):
		contentType = "application/pdf"
	case strings.HasSuffix(filename, ".zip"):
		contentType = "application/zip"

	case strings.HasSuffix(filename, ".xls"):
		contentType = "application/vnd.ms-excel"
	case strings.HasSuffix(filename, ".xlsx"):
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	case strings.HasSuffix(filename, ".html"):
		contentType = "text/html"
	case strings.HasSuffix(filename, ".css"):
		contentType = "text/css"

	case strings.HasSuffix(filename, ".doc"):
		contentType = "application/msword"
	case strings.HasSuffix(filename, ".docx"):
		contentType = "application/msword"

	case strings.HasSuffix(filename, ".png"):
		contentType = "image/png"
	case strings.HasSuffix(filename, ".jpg"),
		strings.HasSuffix(filename, ".jpeg"):
		contentType = "image/jpeg"
	case strings.HasSuffix(filename, ".gif"):
		contentType = "image/gif"
	case strings.HasSuffix(filename, ".svg"):
		contentType = "image/svg+xml"

	case strings.HasSuffix(filename, ".mp4"):
		contentType = "video/mp4"
	case strings.HasSuffix(filename, ".webm"):
		contentType = "video/webm"
	case strings.HasSuffix(filename, ".ogg"):
		contentType = "video/ogg"
	case strings.HasSuffix(filename, ".mp3"):
		contentType = "audio/mp3"
	case strings.HasSuffix(filename, ".wav"):
		contentType = "audio/wav"
	}
	return
}
