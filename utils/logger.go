package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"bitballot/config"
)

func Logger(logPath string) {

	if logPath == "" {
		logPath = config.Get().Path
	} else {
		if _, err := os.Stat(logPath); err != nil {
			if os.IsNotExist(err) {
				os.MkdirAll(logPath, 0777)
			}
		}
	}

	fileName := fmt.Sprintf("/%s.log", filepath.Base(os.Args[0]))
	filePath := fmt.Sprintf(logPath+"logger/%d/%d/%d", time.Now().Year(), time.Now().Month(), time.Now().Day())
	WriteFile(fileName, filePath, []byte(``))

	logfile, err := os.OpenFile(filePath+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		log.Fatalln("Failed to open log file", ":", err)
	}
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(logfile)
}

func WriteFile(fileName, filePath string, fileBytes []byte) bool {
	filePath = config.Get().Path + filePath
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(filePath, 0777)
		} else {
			return false
		}
	}

	if len(fileBytes) > 0 {
		file, err := os.Create(filePath + fileName)
		defer file.Close()
		if err != nil {
			log.Println("Failed Create Error", ":", err)
			return false
		}
		_, err = file.Write(fileBytes)

		if err != nil {
			log.Println("File Write Error: ", err)
			return false
		}
	}
	return true
}
