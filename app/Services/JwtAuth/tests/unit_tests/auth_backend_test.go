package unit_tests

import (
	"github.com/xoxo/crm-x/app/Services/JwtAuth/core/authentication"
	"github.com/xoxo/crm-x/app/Services/JwtAuth/core/redis"
	"github.com/xoxo/crm-x/app/Services/JwtAuth/services/models"
	"github.com/xoxo/crm-x/app/Services/JwtAuth/settings"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	// . "gopkg.in/check.v1"
	"os"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type AuthenticationBackendTestSuite struct{}

var _ = Suite(&AuthenticationBackendTestSuite{})
var t *testing.T

func (s *AuthenticationBackendTestSuite) SetUpSuite(c *C) {
	os.Setenv("GO_ENV", "tests")
	settings.Init()
}

func (suite *AuthenticationBackendTestSuite) TestInitJWTAuthenticationBackend(c *C) {
	authBackend := authentication.InitJWTAuthenticationBackend()
	c.Assert(authBackend, NotNil)
	c.Assert(authBackend.PublicKey, NotNil)
}

func (suite *AuthenticationBackendTestSuite) TestGenerateToken(c *C) {
	authBackend := authentication.InitJWTAuthenticationBackend()
	tokenString, err := authBackend.GenerateToken("1234")

	assert.Nil(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return authBackend.PublicKey, nil
	})

	assert.Nil(t, err)
	assert.True(t, token.Valid)
}

func (suite *AuthenticationBackendTestSuite) TestAuthenticate(c *C) {
	authBackend := authentication.InitJWTAuthenticationBackend()
	user := &models.User{
		Name: "haku",
		Password: "testing",
	}
	c.Assert(authBackend.Authenticate(user), Equals, true)
}

func (suite *AuthenticationBackendTestSuite) TestAuthenticateIncorrectPass(c *C) {
	authBackend := authentication.InitJWTAuthenticationBackend()
	user := models.User{
		Name: "haku",
		Password: "Password",
	}
	c.Assert(authBackend.Authenticate(&user), Equals, false)
}

func (suite *AuthenticationBackendTestSuite) TestAuthenticateIncorrectName(c *C) {
	authBackend := authentication.InitJWTAuthenticationBackend()
	user := &models.User{
		Name: "Haku",
		Password: "testing",
	}
	c.Assert(authBackend.Authenticate(user), Equals, false)
}

func (suite *AuthenticationBackendTestSuite) TestLogout(c *C) {
	authBackend := authentication.InitJWTAuthenticationBackend()
	tokenString, err := authBackend.GenerateToken("1234")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return authBackend.PublicKey, nil
	})
	err = authBackend.Logout(tokenString, token)
	assert.Nil(t, err)

	redisConn := redis.Connect()
	redisValue, err := redisConn.GetValue(tokenString)
	assert.Nil(t, err)
	assert.NotEmpty(t, redisValue)
}

func (suite *AuthenticationBackendTestSuite) TestIsInBlacklist(c *C) {
	authBackend := authentication.InitJWTAuthenticationBackend()
	tokenString, err := authBackend.GenerateToken("1234")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return authBackend.PublicKey, nil
	})
	err = authBackend.Logout(tokenString, token)
	assert.Nil(t, err)

	assert.True(t, authBackend.IsInBlacklist(tokenString))
}

func (suite *AuthenticationBackendTestSuite) TestIsNotInBlacklist(c *C) {
	authBackend := authentication.InitJWTAuthenticationBackend()
	assert.False(t, authBackend.IsInBlacklist("1234"))
}
