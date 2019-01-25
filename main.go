package main

import (
	"fmt"
	"log"
	"os"

	"bitballot/api"
	"bitballot/config"
	"bitballot/utils"
)

func main() {
	utils.Logger("")
	config.Init(nil) //Init Config.yaml
	api.StartRouter()
}

//Start ...
func Start(TIMEZONE, VERSION, COOKIE, DBPATH, OS, OSPATH, ADDRESS string) {
	//OS e.g "ios" or "android"
	//PATH e.g "/sdcard/com.sample.app/"
	var yaml = []byte(fmt.Sprintf(`timezone: %v
version: %v
cookie: %v
db: %v
os: %v
path: %v
address: %v
encryption_keys:
  public: public.pem
  private: private.pem
`, TIMEZONE, VERSION, COOKIE, DBPATH, OS, OSPATH, ADDRESS))

	utils.Logger(OSPATH)
	config.Init(yaml) //Init Config.yaml
	go api.StartRouter()
}

//Stop ...
func Stop() {
	sMessage := "stopping service @ " + config.Get().Address
	println(sMessage)
	log.Println(sMessage)
	os.Exit(1)
}
