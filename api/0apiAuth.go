package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	// "regexp"
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
	// router.Post("/api/signup", middlewares.ThenFunc(apiAuthSignup))
	// router.Post("/api/signup-newsletter", middlewares.ThenFunc(apiSignupNewsletter))
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


/*
func apiAuthSignup(httpRes http.ResponseWriter, httpReq *http.Request) {

	httpRes.Header().Set("Content-Type", "application/json")

	statusMessage := ""
	statusBody := make(map[string]interface{})
	statusCode := http.StatusInternalServerError

	var formStruct struct {
		Username, Password,
		Email, Mobile,

		Fullname, Title,
		Firstname, Lastname,
		Othername, Street, City,
		State, Country, Occupation,
		Image, Referrer string
	}

	err := json.NewDecoder(httpReq.Body).Decode(&formStruct)
	if err != nil {
		statusMessage = "Error Decoding Form Values " + err.Error()
	} else {
		usersList := []database.Users{}

		sqlQuery := "select id from users where email = ?"
		config.Get().Postgres.Select(&usersList, sqlQuery, formStruct.Email)

		if len(usersList) > 0 {
			statusMessage = fmt.Sprintf("Sorry this Email [%s] already exists", formStruct.Email)
		} else {

			sqlQuery = "select id from users where username = ?"
			config.Get().Postgres.Select(&usersList, sqlQuery, formStruct.Username)

			if len(usersList) > 0 {
				statusMessage = fmt.Sprintf("Sorry this Username [%s] already exists", formStruct.Username)
			} else {

				if formStruct.Username == "" {
					statusMessage += "Username " + IsRequired
				}

				if formStruct.Password == "" {
					statusMessage += "Password " + IsRequired
				}

				if formStruct.Firstname == "" {
					statusMessage += "Firstname " + IsRequired
				}

				if formStruct.Lastname == "" {
					statusMessage += "Lastname " + IsRequired
				}

				emailRE := regexp.MustCompile(`^[a-zA-Z0-9][-_.a-zA-Z0-9]*@[a-zA-Z0-9.-]+\.[a-zA-Z0-9.-]+?$`)
				if emailRE.MatchString(formStruct.Email) {
					statusMessage += "Email " + IsRequired
				}

				if strings.HasSuffix(statusMessage, "\n") {
					statusMessage = statusMessage[:len(statusMessage)-2]
				}

				//All Seems Clear, Create New User Now Now
				if statusMessage == "" {
					profile := database.Profiles{}
					profile.Create(
						map[string]interface{}{
							"Workflow":   "enabled",
							"Title":      formStruct.Title,
							"Fullname":   formStruct.Fullname,
							"Firstname":  formStruct.Firstname,
							"Lastname":   formStruct.Lastname,
							"Othername":  formStruct.Othername,
							"Email":      formStruct.Email,
							"Mobile":     formStruct.Mobile,
							"City":       formStruct.City,
							"State":      formStruct.State,
							"Street":     formStruct.Street,
							"Country":    formStruct.Country,
							"Referrer":   formStruct.Referrer,
							"Occupation": formStruct.Occupation,
						})

					user := database.Users{}
					passwordHash, _ := bcrypt.GenerateFromPassword(
						[]byte(formStruct.Password), bcrypt.DefaultCost)
					user.Create(
						map[string]interface{}{
							"FailedMax": 5,
							"IsCustomer":  true,
							"ProfileID": profile.ID,
							"Email":     formStruct.Email,
							"Mobile":    formStruct.Mobile,
							"Username":  formStruct.Username,
							"Code":      formStruct.Username,
							"Title":     fmt.Sprintf("%v - %v", formStruct.Username, formStruct.Email),
							"Workflow":  "pending",
							"Password":  base64.StdEncoding.EncodeToString(passwordHash),
						})

					statusCode = http.StatusOK
					statusMessage = "Please check your email for details"
					// apiClientWelcomeMail(profile)
				}

			}
		}
	}

	tableHits := database.Hits{}
	tableHits.Code = formStruct.Username
	tableHits.Title = fmt.Sprintf("New Client Signup: [%v] - %s", formStruct.Username, statusMessage)

	tableHits.UserAgent = httpReq.UserAgent()
	tableHits.IPAddress = httpReq.RemoteAddr
	tableHits.Workflow = "enabled"
	tableHits.Description = fmt.Sprintf("Fields: %+v", formStruct)
	tableHits.Create(tableHits.ToMap())

	//

	json.NewEncoder(httpRes).Encode(Message{
		Code:    statusCode,
		Message: statusMessage,
		Body:    statusBody,
	})
	// //Send E-Mail
}
*/


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