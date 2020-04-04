package com

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/hatchify/simply"
)

var configName = ".gomurc"

type prRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

// PRResponse returns the value of github's api response
type PRResponse struct {
	HTTPStatus int    `json:"httpStatus,omitempty"`
	URL        string `json:"html_url,omitempty"`

	Errors []PRResponseError `json:"errors,omitempty"`
}

// PRResponseError returned when PR fails creation
type PRResponseError struct {
	Message string `json:"message,omitempty"`
}

// GitAuthObject represents authentication credentials
type GitAuthObject struct {
	User  string `json:"user"`
	Token string `json:"token"`
}

// LoadAuth will Read credentials from disk
func LoadAuth() (authObject GitAuthObject, err error) {
	usr, err := user.Current()
	if err != nil {
		return
	}
	file, err := ioutil.ReadFile(path.Join(usr.HomeDir, configName))
	if err != nil {
		return
	}
	err = json.Unmarshal(file, &authObject)
	return
}

// Save credentials to disk
func (authObject *GitAuthObject) Save() (err error) {
	data, err := json.Marshal(authObject)
	if err != nil {
		return
	}

	usr, err := user.Current()
	if err != nil {
		return
	}

	return ioutil.WriteFile(path.Join(usr.HomeDir, configName), data, os.ModePerm)
}

// Setup configures credentials from user input
func (authObject *GitAuthObject) Setup() (err error) {
	if logLevel <= SILENT {
		err = fmt.Errorf("unable to read credentials. auth token or user name not found")
		return
	}

	// Set username and token
	var text string
	var user string
	var token string
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n( Access Token Instructions @ https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line )")

	// Parse username and token from command line input
	for err == nil && (len(authObject.User) == 0 || len(authObject.Token) == 0) {
		text = strings.TrimSpace(text)

		if len(user) == 0 {
			// Get username first
			if len(text) > 0 {
				user = text
				text = ""
				continue
			}

			fmt.Print("Enter github username: ")

		} else if len(token) == 0 {
			// Get token and save if username set
			if len(text) > 0 {
				token = text
				authObject.User = user
				authObject.Token = token
				fmt.Println("Saving: ", simply.Stringify(authObject))
				authObject.Save()
				return
			}

			fmt.Print("Enter github personal access token: ")
		}

		text, err = reader.ReadString('\n')
	}

	if err != nil {
		Println("Nevermind then... :(")
	}
	return
}
