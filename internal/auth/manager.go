package auth

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	signingKey string
}

type IndividualRequirements struct {
	jwt.StandardClaims
	GUID string `json:"guid"`
}

func CreateObject(signingKey string) (*Authenticator, error) {
	const op = "auth.Authenticator.CreateObjectAuthenticator"

	if signingKey == "" {
		return nil, fmt.Errorf("%s: %w", op, errors.New("empty signingKey"))
	}

	return &Authenticator{signingKey: signingKey}, nil
}

func (m *Authenticator) CreateObjectJWT(data string, ttl time.Duration) (string, error) {
	guid := uuid.New().String()

	claims := IndividualRequirements{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			Subject:   data,
		},
		GUID: guid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString([]byte(m.signingKey))
}

func (m *Authenticator) CreateObjectRefreshToken() (string, error) {
	const op = "auth.Authenticator.CreateObjectRefreshToken"

	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("%x", b), nil
}

func (m *Authenticator) HashToken(token string) ([]byte, error) {
	const op = "auth.Authenticator.HashToken"

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %w", op, err)
	}

	return hashedToken, nil
}

func (m *Authenticator) CompareTokens(providedToken string, hashedToken []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedToken, []byte(providedToken))

	return err == nil
}
