package componenttest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Key struct {
	KID          string
	PublicKeyB64 string
}

type JWTFeature struct {
	privKey      *rsa.PrivateKey
	kid          string
	publicKeyB64 string
}

func NewJWTFeature() *JWTFeature {
	return &JWTFeature{}
}

// EnsureKeys needs to be called in the test setup or reset step of the component tests in the service
// you're testing so that the keys can be set in the AuthConfig.JWTVerificationPublicKeys map and the JWTs can
// be verified by the service under test
func (j *JWTFeature) EnsureKeys() (Key, error) {
	if j.privKey == nil || j.kid == "" || j.publicKeyB64 == "" {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return Key{}, fmt.Errorf("generate viewer RSA key: %w", err)
		}
		if err := priv.Validate(); err != nil {
			return Key{}, fmt.Errorf("validate viewer RSA key: %w", err)
		}

		newKID := uuid.New().String()
		pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			return Key{}, fmt.Errorf("marshal viewer public key: %w", err)
		}

		j.privKey = priv
		j.kid = newKID
		j.publicKeyB64 = base64.StdEncoding.EncodeToString(pubDER)
	}
	return Key{
		KID:          j.kid,
		PublicKeyB64: j.publicKeyB64,
	}, nil
}

func (j *JWTFeature) CreateJWT(email string, groups []string) (string, error) {
	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"sub":            "viewer-sub",
		"token_use":      "access",
		"auth_time":      now,
		"iss":            "https://cognito-idp.eu-west-2.amazonaws.com/eu-west-2_example",
		"exp":            now + 3600,
		"iat":            now,
		"client_id":      "component-test-client",
		"username":       email,
		"cognito:groups": groups,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = j.kid

	signed, err := t.SignedString(j.privKey)
	return signed, err
}
