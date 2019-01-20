package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type configType struct {
	Address, Contract, Path string
}

var Config configType

func configInit() {
	loggerInit()
	viper.SetConfigType("yaml")
	viper.SetDefault("address", "127.0.0.1:8000")

	viper.SetConfigName("config")
	viper.AddConfigPath("./")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("fatal error config file: %s", err.Error())
		return
	}

	Config = configType{}

	Config.Address = viper.GetString("address")
	Config.Contract = viper.GetString("contact")
	Config.Path = viper.GetString("path")
	apiInit()
}

func loggerInit() {
	fileName := fmt.Sprintf("/%s.log", filepath.Base(os.Args[0]))
	filePath := fmt.Sprintf("logger/%d/%d/%d", time.Now().Year(), time.Now().Month(), time.Now().Day())
	writeFile(fileName, filePath, []byte(``))

	logfile, err := os.OpenFile(filePath+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		log.Fatalln("Failed to open log file", ":", err)
	}
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(logfile)
}

func writeFile(fileName, filePath string, fileBytes []byte) bool {
	filePath = Config.Path + filePath
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
