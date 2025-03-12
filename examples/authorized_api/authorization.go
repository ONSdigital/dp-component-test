package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func ZebedeeMustPermitUser(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := zebedeeValidateUser()
		if err != nil {
			if status == 401 {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(status)
		message := "[\"DELETE\", \"READ\", \"CREATE\", \"UPDATE\"]"
		fmt.Fprintf(w, "%s", message)
		handler(w, r)
	}
}

func zebedeeValidateUser() (int, error) {
	type Permissions struct {
		PermissionsMsg string `bson:"message" json:"message"`
	}
	var permissions Permissions
	config := NewConfig()
	response, err := http.Get(config.authorizationServiceUrl + "/userInstancePermissions")
	if err != nil {
		return 500, err
	}
	status := response.StatusCode
	if status == 401 {
		body, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			return status, err
		}
		if len(body) == 0 {
			return status, errors.New("user has not been authorised by zebedee")
		}
		err = json.Unmarshal(body, &permissions)
		if err != nil {
			return status, err
		}
		return status, errors.New(permissions.PermissionsMsg)
	}
	return status, nil
}

func MustAuthorize(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := validateAuth(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

func ZebedeeMustAuthorize(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := zebedeeValidateAuth()
		if err != nil {
			if status == 401 {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(status)
		message := "[\"DELETE\", \"READ\", \"CREATE\", \"UPDATE\"]"
		fmt.Fprintf(w, "%s", message)
		handler(w, r)
	}
}

func zebedeeValidateAuth() (int, error) {
	type Permissions struct {
		PermissionsMsg string `bson:"message" json:"message"`
	}
	var permissions Permissions
	config := NewConfig()
	response, err := http.Get(config.authorizationServiceUrl + "/serviceInstancePermissions")
	if err != nil {
		return 500, err
	}
	status := response.StatusCode
	if status == 401 {
		body, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			return status, err
		}
		if len(body) == 0 {
			return status, errors.New("service has not been authorised by zebedee")
		}
		err = json.Unmarshal(body, &permissions)
		if err != nil {
			return status, err
		}
		return status, errors.New(permissions.PermissionsMsg)
	}
	return status, nil
}

func validateAuth(token string) error {
	if token == "" {
		return errors.New("401 - Unauthorized")
	}
	return nil
}

func MustBeIdentified(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := validateIdentity()
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

func validateIdentity() error {
	type Identity struct {
		Identity string `bson:"identity" json:"identifier"`
	}
	var identity Identity
	config := NewConfig()
	response, err := http.Get(config.authorizationServiceUrl + "/identity")
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("user has not been identified as an admin")
	}
	err = json.Unmarshal(body, &identity)
	if err != nil {
		return err
	}
	if identity.Identity != "admin" {
		return errors.New("user has not been identified as an admin")
	}
	return nil
}
