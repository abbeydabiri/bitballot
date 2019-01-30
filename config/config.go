package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/spf13/viper"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	keySize   = 32
	nonceSize = 24
)

//Config structure
type Config struct {
	Timezone, Version, COOKIE, OS,
	Path, Address, Adminurl string

	Encryption struct {
		Private []byte
		Public  []byte
	}

	Ethereum struct {
		Network, Mnemonic,
		Smartcontract string
	}
	
	Postgres *sqlx.DB
	dbConfig map[string]string
}

var config Config

//Get ...
func Get() *Config {
	return &config
}

//Init ...
func Init(yamlConfig []byte) {

	viper.SetConfigType("yaml")
	viper.SetDefault("address", "127.0.0.1:8080")

	var err error
	if yamlConfig == nil {
		viper.SetConfigName("config")
		viper.AddConfigPath("./")  // optionally look for config in the working directory
		err = viper.ReadInConfig() // Find and read the config file
	} else {
		err = viper.ReadConfig(bytes.NewBuffer(yamlConfig))
	}

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	config = Config{}
	config.OS = viper.GetString("os")
	config.Path = viper.GetString("path")

	config.COOKIE = viper.GetString("cookie")
	config.Address = viper.GetString("address")
	config.Version = viper.GetString("version")
	config.Timezone = viper.GetString("timezone")
	if config.Adminurl = viper.GetString("adminurl"); config.Adminurl == "" {
		config.Adminurl = config.Address
	}

	encrptionKeysMap := viper.GetStringMapString("encryption_keys")
	if encrptionKeysMap != nil {
		config.Encryption.Public, err = Asset(encrptionKeysMap["public"])
		if err != nil {
			log.Fatalf("Error reading public key %v", err)
			return
		}

		config.Encryption.Private, err = Asset(encrptionKeysMap["private"])
		if err != nil {
			log.Fatalf("Error reading private key %v", err)
			return
		}
	}

	//Ethereum Keys
	ethereumMap := viper.GetStringMapString("ethereum")
	if ethereumMap != nil {
		config.Ethereum.Network = ethereumMap["network"]
		config.Ethereum.Mnemonic = strings.TrimSpace(ethereumMap["mnemonic"])
		config.Ethereum.Smartcontract = strings.TrimSpace(ethereumMap["smartcontract"])
	}
	//Ethereum Keys

	//SQL Connection for POSTGRESQL
	if config.dbConfig = viper.GetStringMapString("dbconfig"); len(config.dbConfig) == 5 {
		postgresConn := "host=%s port=%s dbname=%s user=%s password=%s sslmode=disable connect_timeout=1"
		postgresConn = fmt.Sprintf(postgresConn, config.dbConfig["hostname"], config.dbConfig["port"],
			config.dbConfig["database"], config.dbConfig["username"], config.dbConfig["password"])
		config.Postgres, err = sqlx.Open("postgres", postgresConn)
		err = config.Postgres.Ping()
		log.Printf("Connecting Database..")
		if err != nil {
			log.Fatalf("Error Connecting Database %v", err)
			os.Exit(1)
			return
		}
	}
	//SQL Connection for POSTGRESQL
}

func Encrypt(in []byte) (out []byte) {
	key, nonce := keyNounce()
	out = secretbox.Seal(out, in, nonce, key)
	return
}

func Decrypt(in []byte) (out []byte) {
	key, nonce := keyNounce()
	out, _ = secretbox.Open(out, in, nonce, key)
	return
}

func spaceRemove(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func keyNounce() (key *[keySize]byte, nonce *[nonceSize]byte) {
	fullPath := filepath.Dir(os.Args[0])
	fullPath = spaceRemove(fullPath)
	fullPath = strings.Replace(fullPath, "/", "", -1)
	fullPath = strings.Replace(fullPath, "\\", "", -1)

	fullPath = base64.StdEncoding.EncodeToString([]byte(fullPath))
	nPower := int(60 / len(fullPath))
	if len(fullPath) < 60 {
		nCount := 0
		for nPower > nCount {
			fullPath += fullPath
			nCount++
		}
		fullPath = fullPath[0:60]
	}

	key = new([keySize]byte)
	copy(key[:], []byte(fullPath[0:32])[:keySize])

	nonce = new([nonceSize]byte)
	copy(nonce[:], []byte(fullPath[0:32][0:24])[:nonceSize])

	return
}
