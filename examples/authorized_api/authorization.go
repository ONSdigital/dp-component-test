package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

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
		err := zebedeeValidateAuth(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

func zebedeeValidateAuth(token string) error {
	if token == "" {
		type Permissions struct {
			Permissions string `bson:"message" json:"message"`
		}
		var permissions Permissions
		config := NewConfig()
		response, err := http.Get(config.authorizationServiceUrl + "/serviceInstancePermissions")
		if err != nil {
			return err
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		if len(body) == 0 {
			return errors.New("user has not been authorised by zebedee")
		}
		err = json.Unmarshal(body, &permissions)
		if err != nil {
			return err
		}
		return errors.New(permissions.Permissions)
	}
	return nil
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
		return errors.New("User has not been identified as an admin")
	}
	err = json.Unmarshal(body, &identity)
	if err != nil {
		return err
	}
	if identity.Identity != "admin" {
		return errors.New("User has not been identified as an admin")
	}
	return nil
}
