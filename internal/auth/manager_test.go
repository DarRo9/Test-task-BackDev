package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateObjectAuthenticator(t *testing.T) {
	signingKey := "12345"

	_, err := CreateObject(signingKey)
	require.NoError(t, err)
}

func TestCreateObjectAuthenticatorError(t *testing.T) {
	signingKey := ""

	_, err := CreateObject(signingKey)
	require.Error(t, err)
}

func TestCreateObjectJWT(t *testing.T) {
	m := Authenticator{}
	data := "data"
	ttl := time.Duration(time.Duration.Hours(5))

	jwt, err := m.CreateObjectJWT(data, ttl)
	require.NoError(t, err)
	require.NotEmpty(t, jwt)
}

func TestHashToken(t *testing.T) {
	m := Authenticator{}
	token := "4373fbac63c46617971af8e9127cc69dcb8981e3f9b285b08f0b76c6501f7256"

	_, err := m.HashToken(token)
	require.NoError(t, err)
}

func TestCompareTokens(t *testing.T) {
	m := Authenticator{}
	providedToken := "4373fbac63c46617971af8e9127cc69dcb8981e3f9b285b08f0b76c6501f7256"

	hashedToken, _ := m.HashToken(providedToken)

	ok := m.CompareTokens(providedToken, hashedToken)
	require.True(t, ok)
}

func TestCompareTokensError(t *testing.T) {
	m := Authenticator{}
	providedToken := "4373fbac63c46617971af8e9127cc69dcb8981e3f9b285b08f0b76c6501f7256"

	hashedToken := []byte("")

	ok := m.CompareTokens(providedToken, hashedToken)
	require.False(t, ok)
}
