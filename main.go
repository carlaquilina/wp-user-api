package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"wp-user-api/logfile"
	"wp-user-api/util"

	"github.com/gocarina/gocsv"
	"github.com/sogko/go-wordpress"
)

type UserWithError struct {
	WordpressUser wordpress.User
	Error         string
}

var UserExistsError = "This user already exists!"
var SaveFailError = "Could Not Save User To Wordpress!"
var MAIN_URL = "https://carl.customcodesign.com/"
var CURRENT_USER_MAP = make(map[string]wordpress.User)
var DateTime = "2006-01-02_15:04:05"
var timeStamp = time.Now().Format(DateTime)
var logFile, err = logfile.InitialiseLogFile("./logFiles/log" + timeStamp)

func main() {
	if err != nil {
		logFile.LogMessageWithFatal(err.Error())
	}
	logFile.LogMessage("Starting App")
	API_BASE_URL := MAIN_URL + "wp-json/wp/v2"
	USER := "test"
	PASSWORD := "khQTCF000MASzSfXVEHlF"
	logFile.LogMessage("Creating Client for wp")
	client := wordpress.NewClient(&wordpress.Options{
		BaseAPIURL: API_BASE_URL, // example: `http://192.168.99.100:32777/wp-json/wp/v2`
		Username:   USER,
		Password:   PASSWORD,
	})
	//Load Users From WordPress
	LoadCurrentUsers(client, &CURRENT_USER_MAP)
	//Load Data From CSV
	usersFileNameNoExt := "users"
	usersFileExt := ".csv"
	usersCSVPath := "./" + usersFileNameNoExt + usersFileExt
	csvUsers := GetCSV(usersCSVPath)
	err = util.BackUpFile(usersCSVPath, "./backup/"+usersFileNameNoExt+"_"+timeStamp+usersFileExt)
	if err != nil {
		logFile.LogMessageWithFatal(err.Error())
	}
	logFile.LogMessage("backup of file: " + usersCSVPath + " success!")
	//Check if users exist... this function will also remove such users as it will be usless to try and save them to wp
	usersLeft, err := CheckIfUsersExist(&csvUsers, CURRENT_USER_MAP)
	if err != nil {
		logFile.LogMessageWithFatal(err.Error())
	}
	logFile.LogMessage("Check if users exist from csv in wordpress success!")
	err = SaveUsersToWordPress(client, usersLeft)
	if err != nil {
		logFile.LogMessageWithFatal(err.Error())
	}
	logFile.LogMessage("Save users to wordpress success!")
	// logFile.CloseFile("./logFiles/")
}
func SaveUsersToWordPress(client *wordpress.Client, users []wordpress.User) error {
	var usersWithError []UserWithError
	var newUsers []wordpress.User
	for _, user := range users {
		if UserInMap(user.Username, CURRENT_USER_MAP) {
			usersWithError = append(usersWithError, UserWithError{WordpressUser: user, Error: SaveFailError + ", " + UserExistsError})
		} else {
			newUser, err := CreateUser(&user, client)
			if err != nil {
				usersWithError = append(usersWithError, UserWithError{WordpressUser: user, Error: err.Error()})

			} else {
				AddUserToUserMap(user, &CURRENT_USER_MAP)
				newUsers = append(newUsers, *newUser)
			}
		}
	}
	if len(newUsers) > 0 {
		//save the new users to file for reference
		usersCsv, err := ConvertUsersToCSV(newUsers)
		if err != nil {
			return err
		}
		err = SaveStringToFile(usersCsv, "./saving/success/SavedUsers"+timeStamp+".csv")
		if err != nil {
			return err
		}
	}
	if len(usersWithError) > 0 {
		//save the users with errors to file for reference
		usersCsv, err := ConvertUsersWithErrorsToCSV(usersWithError)
		if err != nil {
			return err
		}
		err = SaveStringToFile(usersCsv, "./saving/fail/ErrorSavingUsers"+timeStamp+".csv")
		if err != nil {
			return err
		}
	}
	return nil
}
func CheckIfUsersExist(csvUsers *[]wordpress.User, currentUsers map[string]wordpress.User) ([]wordpress.User, error) {
	var usersWithErrors []UserWithError
	var newCsvUsers []wordpress.User
	for _, csvUser := range *csvUsers {
		//users exist
		if UserInMap(csvUser.Username, currentUsers) {
			//push in var with error
			usersWithErrors = append(usersWithErrors, UserWithError{WordpressUser: csvUser, Error: UserExistsError})
		} else {
			newCsvUsers = append(newCsvUsers, csvUser)
		}
	}
	// csvUsers = &newCsvUsers
	if len(usersWithErrors) > 0 {
		csvContent, err := ConvertUsersWithErrorsToCSV(usersWithErrors)
		if err != nil {
			return newCsvUsers, err
		}
		//save all users with their errors in file, attach a date and time to the filename
		err = SaveStringToFile(csvContent, "./existing/CheckError"+timeStamp+".csv")
		if err != nil {
			return newCsvUsers, err
		}
	}
	return newCsvUsers, nil
}
func CreateUser(u *wordpress.User, client *wordpress.Client) (*wordpress.User, error) {
	newUser, resp, body, err := client.Users().Create(u)
	if err != nil {
		err := fmt.Sprintf("Should not return error: %v", err.Error())
		return newUser, errors.New(err)
	}
	if body == nil {
		err := fmt.Sprintf("body should not be nil")
		return newUser, errors.New(err)
	}
	if resp.StatusCode != http.StatusCreated {
		err := fmt.Sprintf("Expected 201 Created, got %v", resp.Status)
		return newUser, errors.New(err)
	}
	if newUser == nil {
		err := fmt.Sprintf("newUser should not be nil")
		return newUser, errors.New(err)
	}
	return newUser, nil
}
func LoadCurrentUsers(client *wordpress.Client, userMap *map[string]wordpress.User) []wordpress.User {
	logFile.LogMessage("Loading current users!")
	users, resp, body, err := client.Users().List(nil)
	if err != nil {
		logFile.LogMessageWithLog(fmt.Sprintf("Should not return error: %v", err.Error()))
	}
	if body == nil {
		logFile.LogMessageWithLog("body should not be nil")
	}
	if resp.StatusCode != http.StatusOK {
		logFile.LogMessageWithLog(fmt.Sprintf("Expected 201 Created, got %v", resp.Status))
	}
	for _, user := range users {
		user.Username = strings.TrimPrefix(user.Link, MAIN_URL+"author/")
		user.Username = strings.TrimSuffix(user.Username, "/")
		AddUserToUserMap(user, userMap)
	}
	return users
}
func AddUserToUserMap(user wordpress.User, userMap *map[string]wordpress.User) {

	_userMap := *userMap
	_userMap[user.Username] = user
	userMap = &_userMap
}
func PrintUserMap(userMap map[string]wordpress.User) {
	for k, _ := range userMap {
		logFile.LogMessageWithLog(k)
	}
}
func PrintMarshalIndentMap(userMap map[string]wordpress.User) {
	b, err := json.MarshalIndent(userMap, "", "  ")
	if err != nil {
		logFile.LogMessageWithLog(fmt.Sprintf("error:", err))
	}
	fmt.Print(string(b))
}
func UserInMap(key string, usersMap map[string]wordpress.User) bool {
	if _, ok := usersMap[key]; ok {
		return true
	}
	return false
}

func GetCSV(pathToFile string) []wordpress.User {
	logFile.LogMessage("Getting users from csv file: " + pathToFile)
	if !util.FileExists(pathToFile) {
		logFile.LogMessageWithFatal("This file does not exist: " + pathToFile)
	}
	clientsFile, err := os.OpenFile(pathToFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		logFile.LogMessageWithFatal(err.Error())
	}
	defer clientsFile.Close()

	clients := []wordpress.User{}

	if err := gocsv.UnmarshalFile(clientsFile, &clients); err != nil { // Load clients from file
		logFile.LogMessageWithLog(err.Error())
	}
	return clients
}

func ConvertUsersWithErrorsToCSV(usersWithErrors []UserWithError) (string, error) {
	csvContent, err := gocsv.MarshalString(&usersWithErrors)
	if err != nil {
		return "", err
	}
	return csvContent, nil
}
func ConvertUsersToCSV(users []wordpress.User) (string, error) {
	csvContent, err := gocsv.MarshalString(&users)
	if err != nil {
		return "", err
	}
	return csvContent, nil
}
func SaveStringToFile(data, pathToFile string) error {
	f, err := os.Create(pathToFile)
	if err != nil {
		return err
	}
	_, err = f.WriteString(data)
	if err != nil {
		f.Close()
		return err
	}
	// fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}
