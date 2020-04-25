package com

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"strings"
)

var configName = ".gomurc"

type secretRequest struct {
	Encrypted string `json:"encrypted_value,omitempty"`

	PublicKey string `json:"key,omitempty"`
	KeyID     string `json:"key_id,omitempty"`

	HTTPStatus int               `json:"httpStatus,omitempty"`
	Errors     []PRResponseError `json:"errors,omitempty"`
}

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

// Encrypt will salt a secret using sodium lib, and return the encrypted value
func (authObject *GitAuthObject) Encrypt(secret, key string) (encrypted string, err error) {
	// TODO: Sodium encrypt https://help.github.com/actions/automating-your-workflow-with-github-actions/creating-and-using-encrypted-secrets

	return
}

// GetPublicKey will set public key and key id for encrypting secrets
func (authObject *GitAuthObject) GetPublicKey(goURL string) (id, key string, err error) {
	// Get git host
	comps := strings.Split(goURL, "/")
	switch comps[0] {
	case "github.com":
		// Supported
	default:
		// Not Supported
		err = fmt.Errorf("%s currently not supported for pull requests", comps[0])
		return
	}

	// GitHub api url
	apiURL := "https://api." + comps[0]
	resource := "/repos/" + strings.Join(comps[1:], "/") + "/actions/secrets/public-key"

	u, err := url.ParseRequestURI(apiURL)
	if err != nil {
		err = fmt.Errorf("Unable to parse url %s", apiURL)
		return
	}

	// Check auth token
	if err != nil || len(authObject.User) == 0 || len(authObject.Token) == 0 {
		// Get new creds
		err = fmt.Errorf("Unable to parse github username and token")
		return
	}

	// Make request
	u.Path = resource
	urlStr := u.String()
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", "token "+authObject.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	// Execute Request
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	// Read response
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	resp.Body.Close()
	payload := &secretRequest{}
	err = json.Unmarshal(body, payload)

	// Return status
	payload.HTTPStatus = resp.StatusCode
	if payload.HTTPStatus >= 300 {
		err = fmt.Errorf("Http error %d", payload.HTTPStatus)
		if len(payload.Errors) > 0 {
			err = fmt.Errorf("Http Error %d: %s", payload.HTTPStatus, payload.Errors[0].Message)
			return
		}
	}

	return
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
				if err = authObject.Save(); err != nil {
					fmt.Println("Error saving credentials :(\n", err)
				} else {
					fmt.Println("Saved Credentials!")
				}
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
