package main

import (
	"errors"
	"net/http"
)

func MustAuthorize(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := validateAuth("some token")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

func validateAuth(token string) error {
	config := NewConfig()
	response, err := http.Get(config.authorizationServiceUrl + "/identity")
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errors.New("401 - Unauthorized")
	}

	return nil
}
