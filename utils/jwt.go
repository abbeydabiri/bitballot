package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitballot/config"

	jwt "github.com/dgrijalva/jwt-go"
)

func VerifyJWT(httpRes http.ResponseWriter, httpReq *http.Request) (claims jwt.MapClaims) {

	if monsterCookie, err := httpReq.Cookie(config.Get().COOKIE); err == nil {
		claims = ValidateJWT(monsterCookie.Value)

		if claims == nil {
			cookieMonster := &http.Cookie{
				Name: config.Get().COOKIE, Value: "deleted", Path: "/",
				Expires: time.Now().Add(-(time.Hour * 24 * 30 * 12)), // set the expire time
			}
			http.SetCookie(httpRes, cookieMonster)
		}
	}

	return
}

func ValidateJWT(jwtToken string) (claims jwt.MapClaims) {
	token, _ := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		publicKey, _ := jwt.ParseRSAPublicKeyFromPEM(config.Get().Encryption.Public)
		return publicKey, nil
	})

	if token != nil {
		jwtClaims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			if jwtClaims["claims"] != nil {
				base64Bytes, err := base64.StdEncoding.DecodeString(
					jwtClaims["claims"].(string))
				if err != nil {
					log.Println("error: " + err.Error())
					return
				}
				if base64Bytes == nil {
					log.Println("base64Bytes is nil")
					return
				}
				byteClaims := config.Decrypt(base64Bytes)
				json.Unmarshal(byteClaims, &claims)
				if claims["exp"] != nil {
					claims["exp"] = int64(claims["exp"].(float64))
				}
			}
		}
	}
	return
}

//Turn user details into a hashed token that can be used to recognize the user in the future.
func GenerateJWT(claims jwt.MapClaims) (token string, err error) {

	//create new claims with encrypted data
	jwtClaims := jwt.MapClaims{}
	byteClaims, _ := json.Marshal(claims)
	jwtClaims["claims"] = config.Encrypt(byteClaims)
	if claims["exp"] != nil {
		jwtClaims["exp"] = claims["exp"].(int64)
	}

	// create a signer for rsa 256
	t := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), jwtClaims)
	pub, err := jwt.ParseRSAPrivateKeyFromPEM(config.Get().Encryption.Private)
	if err != nil {
		return
	}
	token, err = t.SignedString(pub)
	if err != nil {
		return
	}
	return
}
