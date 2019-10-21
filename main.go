package main

import (
	"encoding/json"
	"fmt"
	"go-wordpress-master"
	"log"
	"net/http"
	"strings"
)

var MAIN_URL = "https://carl.customcodesign.com/"
var CURRENT_USER_MAP = make(map[string]wordpress.User)

func main() {
	API_BASE_URL := MAIN_URL + "wp-json/wp/v2"
	USER := "test"
	PASSWORD := "khQTCF000MASzSfXVEHlF"
	client := wordpress.NewClient(&wordpress.Options{
		BaseAPIURL: API_BASE_URL, // example: `http://192.168.99.100:32777/wp-json/wp/v2`
		Username:   USER,
		Password:   PASSWORD,
	})
	GetCurrentUsers(client, &CURRENT_USER_MAP)
	log.Print("userMap before adding user: ")
	PrintMarshalIndentMap(CURRENT_USER_MAP)
	//the users should be retrieved from a csv or something of the sort
	u := wordpress.User{
		Username: "go-wordpress-test-user1",
		Email:    "go-wordpress-test-user1@email.com",
		Name:     "go-wordpress-test-user1",
		Slug:     "go-wordpress-test-user1",
		Password: "password",
	}
	newUser := CreateUser(u, client)
	AddUserToUserMap(*newUser, &CURRENT_USER_MAP)
	log.Print("userMap after adding user: ")
	PrintMarshalIndentMap(CURRENT_USER_MAP)
}

func CreateUser(u wordpress.User, client *wordpress.Client) *wordpress.User {
	//if the user is not in the map then try and add the user to wordpress
	if UserInMap(u.Username, CURRENT_USER_MAP) {
		log.Println("user with this username already exists")
		return &wordpress.User{}
	}
	newUser, resp, body, err := client.Users().Create(&u)
	if err != nil {
		log.Printf("Should not return error: %v", err.Error())
	}
	if body == nil {
		log.Printf("body should not be nil")
	}
	if resp.StatusCode != http.StatusCreated {
		log.Printf("Expected 201 Created, got %v", resp.Status)
	}
	if newUser == nil {
		log.Printf("newUser should not be nil")
	}
	return newUser
}

func GetCurrentUsers(client *wordpress.Client, userMap *map[string]wordpress.User) []wordpress.User {
	users, resp, body, err := client.Users().List(nil)
	if err != nil {
		log.Printf("Should not return error: %v", err.Error())
	}
	if body == nil {
		log.Printf("body should not be nil")
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Expected 201 Created, got %v", resp.Status)
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
		log.Println(k)
	}
}
func PrintMarshalIndentMap(userMap map[string]wordpress.User) {
	b, err := json.MarshalIndent(userMap, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))
}
func UserInMap(key string, usersMap map[string]wordpress.User) bool {
	if _, ok := usersMap[key]; ok {
		return true
	}
	return false
}
