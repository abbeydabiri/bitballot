package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"regexp"

	"bitballot/config"
	"bitballot/database"
	"bitballot/utils"

	"github.com/justinas/alice"
)

func apiHandlerImport(middlewares alice.Chain, router *Router) {
	router.Post("/api/import", middlewares.ThenFunc(apiImport))
}

func apiImport(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRes.Header().Set("Content-Type", "application/json")

	var statusBody interface{}
	statusCode := http.StatusInternalServerError
	statusMessage := ""

	if claims := utils.VerifyJWT(httpRes, httpReq); claims == nil {
		statusBody = map[string]string{"Redirect": "/"}
	} else {

		var formStruct struct{ File, Bucket string }
		err := json.NewDecoder(httpReq.Body).Decode(&formStruct)
		if err != nil {
			statusMessage = "Error Decoding Form Values: " + err.Error()
		} else {

			if len(strings.Split(formStruct.File, "base64,")) != 2 {
				statusMessage = "Invalid File uploaded"
			} else {

				base64Bytes, err := base64.StdEncoding.DecodeString(
					strings.Split(formStruct.File, "base64,")[1])
				if base64Bytes != nil && err == nil {

					var tablesList []map[string]interface{}
					switch strings.ToLower(formStruct.Bucket) {
					case "recipients":
						nextSequence := uint64(0)
						dealerList := string(base64Bytes)
						dealerList = strings.Replace(dealerList, "\r", "", -1)
						dealerList = strings.Replace(dealerList, "\n\n", "\n", -1)
						sliceRow := strings.Split(dealerList, "\n")

						if len(sliceRow) > 0 {

							if len(sliceRow) > 15000 {
								statusMessage = "Please limit import to 15000 lines"
							} else {

								RecipientsID := make(map[uint64]string)

								//Fetch All Profiles and Sort by Email and Mobile
								tableProfileList := []database.Profiles{}
								tableProfileEmail := make(map[string]uint64)
								tableProfileMobile := make(map[string]uint64)
								tableProfileUpdate := make(map[uint64]database.Profiles)

								config.Get().Postgres.Select(&tableProfileList, "select * from profiles")
								for _, tableProfile := range tableProfileList {
									if claims["IsAdmin"] == nil || !claims["IsAdmin"].(bool) {
										if tableProfile.Createdby != uint64(claims["ID"].(float64)) {
											continue
										}
									}

									if tableProfile.Email != "" {
										tableProfileUpdate[tableProfile.ID] = tableProfile
										tableProfileEmail[tableProfile.Email] = tableProfile.ID
									}

									if tableProfile.Mobile != "" {
										tableProfileUpdate[tableProfile.ID] = tableProfile
										tableProfileMobile[tableProfile.Mobile] = tableProfile.ID
									}
								}
								//Fetch All Profiles and Sort by Email and Mobile

								if statusMessage == "" {
									var lContinue bool

									for _, stringCols := range sliceRow {
										lContinue = true
										sliceCols := strings.Split(strings.TrimSpace(stringCols), ",")

										tableProfile := database.Profiles{}
										for index, value := range sliceCols {
											if index > 2 {
												continue
											}

											fieldName := ""
											value = strings.TrimSpace(value)

											mobileRE := regexp.MustCompile(`^\d{10,}$`)
											if mobileRE.MatchString(value) {
												if fieldName == "" {
													fieldName = "Mobile"
												}
											}

											emailRE := regexp.MustCompile(`^[a-zA-Z0-9][-_.a-zA-Z0-9]*@[a-zA-Z0-9.-]+\.[a-zA-Z0-9.-]+?$`)
											if emailRE.MatchString(value) {
												if fieldName == "" {
													fieldName = "Email"
												}
											}

											if fieldName == "" {
												fieldName = "Fullname"
											}

											switch fieldName {
											case "Fullname":
												if tableProfile.Fullname == "" {
													tableProfile.Fullname = strings.TrimSpace(value)
												}
											case "Mobile":
												if tableProfile.Mobile == "" {
													tableProfile.Mobile = strings.TrimSpace(value)
												}
											case "Email":
												if tableProfile.Email == "" {
													tableProfile.Email = strings.ToLower(strings.TrimSpace(value))
												}
											}
										}

										addRecipientsContact := func(tableProfile database.Profiles) (sDetails string) {

											if tableProfile.Fullname != "" {
												sDetails = tableProfile.Fullname
											}

											if tableProfile.Email != "" {
												if sDetails != "" {
													sDetails += ", "
												}

												sDetails += tableProfile.Email
											}

											if tableProfile.Mobile != "" {
												if sDetails != "" {
													sDetails += ", "
												}
												sDetails += tableProfile.Mobile
											}
											return
										}

										//Check if Email exists in User List and Use that instead of creating a new one
										if tableProfile.Email != "" && tableProfileEmail[tableProfile.Email] > uint64(0) {
											userDetail := tableProfileUpdate[tableProfileEmail[tableProfile.Email]]
											RecipientsID[userDetail.ID] = addRecipientsContact(userDetail)
											lContinue = false
										}

										//Check if Email exists in User List and Use that instead of creating a new one
										if tableProfile.Mobile != "" && tableProfileMobile[tableProfile.Mobile] > uint64(0) {
											userDetail := tableProfileUpdate[tableProfileMobile[tableProfile.Mobile]]
											RecipientsID[userDetail.ID] = addRecipientsContact(userDetail)
											lContinue = false
										}

										if tableProfile.Fullname == "" {
											if tableProfile.Email != "" {
												tableProfile.Fullname = tableProfile.Email
											} else if tableProfile.Mobile != "" {
												tableProfile.Fullname = tableProfile.Mobile
											} else {
												lContinue = false
											}
										}

										if tableProfile.Email == "" && tableProfile.Mobile == "" {
											lContinue = false
										}

										if lContinue {
											//Create New User and Add to RecipientsID
											tableProfile.Createdby = uint64(claims["ID"].(float64))

											//Get Current Sequence, auto increment

											// tableProfile.ID = nextSequence
											nextSequence++
											go tableProfile.Create(tableProfile.ToMap())
											RecipientsID[nextSequence] = addRecipientsContact(tableProfile)

											if tableProfileUpdate[nextSequence].ID == uint64(0) {
												tableProfileUpdate[nextSequence] = tableProfile
												if tableProfile.Email != "" {
													tableProfileEmail[tableProfile.Email] = nextSequence
												}

												if tableProfile.Email != "" {
													tableProfileMobile[tableProfile.Mobile] = nextSequence
												}
											}

										}

									}

									statusBody = RecipientsID
									statusCode = http.StatusOK
									statusMessage = fmt.Sprintf("%v Unique Contacts Idenitfied", len(RecipientsID))
								}
							}
						}

					case "profiles":
						if !strings.HasPrefix(formStruct.File, "data:application/csv;base64,") &&
							!strings.HasPrefix(formStruct.File, "data:text/csv;base64,") {
							statusMessage = "Error File is Not CSV"
							break
						}

						pipeCSV := string(base64Bytes)
						pipeCSV = strings.Replace(pipeCSV, "\r", "", -1)
						pipeCSV = strings.Replace(pipeCSV, "\n\n", "\n", -1)
						sliceRow := strings.Split(pipeCSV, "\n")

						if len(sliceRow) > 15000 {
							statusMessage = "Please limit import to 15000 lines"
							break
						}

						var profileList []map[string]interface{}
						for index, stringCols := range sliceRow {
							if index == 0 {
								continue
							}

							table := database.Profiles{}
							sliceCols := strings.Split(strings.TrimSpace(stringCols), "|")
							for index, value := range sliceCols {
								value = strings.TrimSpace(value)
								switch index {
								case 0:
									nameList := strings.Split(value, " ")
									table.Lastname = nameList[0]
									if len(nameList) > 1 {
										table.Firstname = nameList[1]
									}
									if len(nameList) > 2 {
										table.Othername += nameList[2]
									}
									table.Fullname = fmt.Sprintf("%s %s", value, table.Lastname)
								}
							}
							table.Fullname = strings.TrimSpace(table.Fullname)
							if table.Fullname != "" {
								profileList = append(profileList, table.ToMap())
							}
						}

						if len(profileList) > 0 {
							sqlBulkInsert := database.SQLBulkInsert(&database.Profiles{}, profileList)
							config.Get().Postgres.Exec(sqlBulkInsert)
						}
						statusCode = http.StatusOK
						statusMessage = fmt.Sprintf("%v Profiles Imported", len(profileList))

					default:
						if database.AllTables[formStruct.Bucket] != nil {
							if strings.HasPrefix(formStruct.File, "data:application/json;base64,") && statusMessage == "" {
								err := json.Unmarshal(base64Bytes, &tablesList)
								if err == nil {
									sqlBulkInsert := database.SQLBulkInsert(database.AllTables[formStruct.Bucket], tablesList)
									go func() {
										config.Get().Postgres.Exec(sqlBulkInsert)
									}()
									statusCode = http.StatusOK
									statusMessage = fmt.Sprintf("%v Records Imported", len(tablesList))

								} else {
									statusMessage = err.Error()
								}
							} else {
								statusMessage = "Error File is Not Valid JSON"
							}
						} else {
							statusMessage = "Bucket not Setup for Import"
						}
					}

				} else {
					statusMessage = err.Error()
				}
			}
		}
	}

	json.NewEncoder(httpRes).Encode(Message{
		Code:    statusCode,
		Body:    statusBody,
		Message: statusMessage,
	})
}
