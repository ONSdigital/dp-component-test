package main

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

type Config struct {
	AuthConfig AuthConfig
}

type AuthConfig struct {
	JWTVerificationPublicKeys map[string]string
}

func NewConfig() *Config {
	return &Config{
		AuthConfig: AuthConfig{JWTVerificationPublicKeys: map[string]string{}},
	}
}

func ExampleHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(201)
	fmt.Fprintf(w, "accepted")
}

func NewRouter(cfg *Config) http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/checkjwt", checkJWT(cfg)).Methods(http.MethodGet)

	return router
}

func checkJWT(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)

		keyFunc := func(token *jwt.Token) (interface{}, error) {
			kid, _ := token.Header["kid"].(string)
			pubKeyB64 := config.AuthConfig.JWTVerificationPublicKeys[kid]
			pubKeyDER, _ := base64.StdEncoding.DecodeString(pubKeyB64)
			pubKey, _ := x509.ParsePKIXPublicKey(pubKeyDER)
			return pubKey, nil
		}

		token, err := jwt.Parse(tokenString, keyFunc)
		if err != nil || !token.Valid {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	cfg := NewConfig()
	server := &http.Server{
		Addr:              ":10000",
		Handler:           NewRouter(cfg),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
