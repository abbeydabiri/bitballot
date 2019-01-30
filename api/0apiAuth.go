package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"regexp"
	// "strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/justinas/alice"
	"golang.org/x/crypto/bcrypt"

	"bitballot/config"
	"bitballot/database"
	"bitballot/utils"
)

func apiHandlerAuth(middlewares alice.Chain, router *Router) {
	router.Post("/api/login", middlewares.ThenFunc(apiAuthLogin))
	// router.Post("/api/forgot", middlewares.ThenFunc(apiAuthForgot))
	// router.Post("/api/signup-newsletter", middlewares.ThenFunc(apiSignupNewsletter))

	router.Post("/api/register", middlewares.ThenFunc(apiAuthRegister))
	router.Post("/api/checkaddress", middlewares.ThenFunc(apiAuthCheckAddress))
}

func apiAuthLogin(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRes.Header().Set("Content-Type", "application/json")

	var formStruct struct {
		Username, Password, Path string
	}

	statusBody := make(map[string]interface{})
	statusCode := http.StatusInternalServerError
	statusMessage := "Invalid Username or Password"

	err := json.NewDecoder(httpReq.Body).Decode(&formStruct)
	if err == nil {

		usersList := []database.Users{}
		sqlQuery := "select * from users where workflow = 'enabled' and username = $1 "
		config.Get().Postgres.Select(&usersList, sqlQuery, formStruct.Username)

		if len(usersList) == 1 {
			lValid := true
			user := usersList[0]
			userMap := make(map[string]interface{})
			userMap["ID"] = user.ID

			passwordHash, _ := base64.StdEncoding.DecodeString(user.Password)
			if err = bcrypt.CompareHashAndPassword(passwordHash, []byte(formStruct.Password)); err != nil {
				lValid = true
			} else {
				formStruct.Password = ""
			}

			if user.Workflow != "enabled" && user.Workflow != "active" && formStruct.Username != "root" {
				lValid = true
			}

			if !lValid {
				if formStruct.Username != "root" {
					user.Failed++
					if user.FailedMax <= user.Failed {
						user.Workflow = "blocked"
						user.Failed = user.FailedMax
						statusMessage = fmt.Sprintf("User account blocked - too many failed logins")
					} else {
						statusMessage = fmt.Sprintf("%v attempts left", user.FailedMax-user.Failed)
					}
					user.Update(user.ToMap())
				}
			} else {
				// All Seems Clear, Validate User Password and Generate Token
				userMap["Failed"] = uint64(0)
				user.Update(userMap)

				// set our claims
				jwtClaims := jwt.MapClaims{}
				jwtClaims["ID"] = user.ID
				jwtClaims["Username"] = user.Username
				jwtClaims["Email"] = user.Email
				jwtClaims["Mobile"] = user.Mobile
				jwtClaims["IsAdmin"] = true
				statusBody["Redirect"] = "/admin"

				if statusBody["Redirect"] != nil {
					cookieExpires := time.Now().Add(time.Hour * 24 * 14) // set the expire time
					jwtClaims["exp"] = cookieExpires.Unix()

					if jwtToken, err := utils.GenerateJWT(jwtClaims); err == nil {
						cookieMonster := &http.Cookie{
							Name: config.Get().COOKIE, Value: jwtToken, Expires: cookieExpires, Path: "/",
						}
						http.SetCookie(httpRes, cookieMonster)
						httpReq.AddCookie(cookieMonster)

						statusCode = http.StatusOK
						statusMessage = "User Verified"
					}
				}
				//All Seems Clear, Validate User Password and Generate Token
			}
		}
	} else {
		println(err.Error())
	}

	json.NewEncoder(httpRes).Encode(Message{
		Code:    statusCode,
		Message: statusMessage,
		Body:    statusBody,
	})
}



func apiAuthRegister(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRes.Header().Set("Content-Type", "application/json")

	message := new(Message)
	message.Message = ""
	message.Code = http.StatusInternalServerError

	var formStruct struct {
		Firstname, Lastname,
		Email, Mobile,
		Code string
	}

	if err := json.NewDecoder(httpReq.Body).Decode(&formStruct); err != nil {
		message.Message += "Error Decoding Form Values " + err.Error()
		json.NewEncoder(httpRes).Encode(message)
		return
	}

	if formStruct.Code == "" {
		message.Message += "Address " + IsRequired
	}

	if formStruct.Firstname == "" {
		message.Message += "Firstname " + IsRequired
	}

	if formStruct.Lastname == "" {
		message.Message += "Lastname " + IsRequired
	}

	if formStruct.Mobile == "" {
		message.Message += "Mobile " + IsRequired
	}

	emailRE := regexp.MustCompile(`^[a-zA-Z0-9][-_.a-zA-Z0-9]*@[a-zA-Z0-9.-]+\.[a-zA-Z0-9.-]+?$`)
	if emailRE.MatchString(formStruct.Email) {
		message.Message += "Email " + IsRequired
	}


	if message.Message != "" {
		json.NewEncoder(httpRes).Encode(message)
		return
	}

	nCountEmail := 0
	sqlQueryEmail := "select count(id) from profiles where email = ?"
	config.Get().Postgres.Select(&nCountEmail, sqlQueryEmail, formStruct.Email)
	if  nCountEmail > 0 {
		message.Message += "Email is already taken"
		json.NewEncoder(httpRes).Encode(message)
		return
	}

	nCountMobile := 0
	sqlQueryMobile := "select count(id) from profiles where email = ?"
	config.Get().Postgres.Select(&nCountMobile, sqlQueryMobile, formStruct.Mobile)
	if  nCountMobile > 0 {
		message.Message += "Mobile is already taken"
		json.NewEncoder(httpRes).Encode(message)
		return
	}

	profile := database.Profiles{}
	profile.Create(
		map[string]interface{}{
			"Workflow":   "pending",

			"Code":       formStruct.Code,
			"Fullname":   formStruct.Firstname+" "+formStruct.Lastname,
			"Firstname":  formStruct.Firstname,
			"Lastname":   formStruct.Lastname,
			
			"Email":      formStruct.Email,
			"Mobile":     formStruct.Mobile,
		})

	message.Code = http.StatusOK
	message.Message = "You have been registered"
	json.NewEncoder(httpRes).Encode(message)
}

func apiAuthCheckAddress(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRes.Header().Set("Content-Type", "application/json")

	message := new(Message)
	message.Message = ""
	message.Code = http.StatusInternalServerError

	var formStruct struct {
		Code string
	}

	if err := json.NewDecoder(httpReq.Body).Decode(&formStruct); err != nil {
		message.Message += "Error Decoding Form Values " + err.Error()
		json.NewEncoder(httpRes).Encode(message)
		return
	}

	if formStruct.Code == "" {
		message.Message += "Address " + IsRequired
		json.NewEncoder(httpRes).Encode(message)
	}

	
	nCount := 0
	sqlQuery := "select count(id) from profiles where code = ?"
	config.Get().Postgres.Select(&nCount, sqlQuery, formStruct.Code)
	registered := map[string]bool{"registered":false}
	if  nCount > 0 {
		registered["registered"] = true
	}
	message.Body = registered
	message.Code = http.StatusOK
	json.NewEncoder(httpRes).Encode(message)
}

/*
func apiAuthForgot(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRes.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusOK
	statusMessage := "If Email Exists a Password Reset Link will be sent"

	user := database.Users{}
	err := json.NewDecoder(httpReq.Body).Decode(&user)
	if err == nil {

		sqlQuery := "select username from users where email = $1 limit 1"
		config.Get().Postgres.Get(&user, sqlQuery, user.Email)

		if user.Email != "" {
			//All Seems Clear, Generate Password Reset Activation Link and Mail User

			activation := database.Activations{}
			activationCode := utils.RandomString(128)
			activation.Create(
				map[string]interface{}{
					"Type":       "reset",
					"UserID":     user.ID,
					"Code":       activationCode,
					"Expirydate": time.Now().Add(+(time.Minute * 15)).Format(utils.TimeFormat),
				})

			var mailTemplateValues struct{ Email, Username, ResetLink string }
			mailTemplateValues.Email = user.Email
			mailTemplateValues.Username = user.Username
			mailTemplateValues.ResetLink = activationCode

			emailStruct := utils.Email{}
			emailStruct.To = user.Email
			utils.SendNewsletterEmail(emailStruct, mailTemplateValues, "forgot-password")

			//All Seems Clear, Generate Password Reset Activation Link and Mail User
		}
	}

	json.NewEncoder(httpRes).Encode(Message{
		Code:    statusCode,
		Message: statusMessage,
	})
}
*/