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
