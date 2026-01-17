package handler_test

import (
	"net/http"
	"testing"

	domainuser "github.com/Rusich90/gophermart.git/internal/domain/user"
	"github.com/Rusich90/gophermart.git/internal/http/dto"
	"github.com/Rusich90/gophermart.git/internal/http/handler"
	"github.com/Rusich90/gophermart.git/internal/repository"
	"github.com/Rusich90/gophermart.git/internal/service"
	"github.com/Rusich90/gophermart.git/internal/testutils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuthTestSuite struct {
	suite.Suite
	db        *pgxpool.Pool
	userRepo  domainuser.UserRepo
	httpSetup *testutils.HTTPTestSetup
	jwtSecret []byte
	logger    *zap.Logger
}

func (s *AuthTestSuite) SetupSuite() {
	globalSuite, err := testutils.GetGlobalTestSuite()
	s.Require().NoError(err)
	
	s.db = globalSuite.DB
	s.userRepo = repository.NewUserRepo(s.db)
	s.jwtSecret = []byte("test_secret")
	s.logger = globalSuite.Logger

	s.startServer()
}

func (s *AuthTestSuite) startServer() {
	authService := service.NewAuthService(s.userRepo, s.jwtSecret)
	authHandler := handler.NewAuthHandler(authService, s.logger)

	s.httpSetup = testutils.StartHTTPServer(func(r *gin.Engine) {
		r.POST("/register", authHandler.Register)
		r.POST("/login", authHandler.Login)
	})
}

func (s *AuthTestSuite) TearDownSuite() {
	if s.httpSetup != nil {
		s.httpSetup.Close()
	}
}

func (s *AuthTestSuite) SetupTest() {
	s.truncateTables()
}

func (s *AuthTestSuite) truncateTables() {
	err := testutils.TruncateTables(s.db)
	s.Require().NoError(err, "Failed to truncate tables")
}

func (s *AuthTestSuite) performRequest(method, url string, body interface{}) (*http.Response, error) {
	return s.httpSetup.PerformRequest(method, url, body)
}

func (s *AuthTestSuite) registerUser(login, password string) (*http.Response, error) {
	req := dto.AuthRequest{Login: login, Password: password}
	return s.performRequest("POST", "/register", req)
}

func (s *AuthTestSuite) loginUser(login, password string) (*http.Response, error) {
	req := dto.AuthRequest{Login: login, Password: password}
	return s.performRequest("POST", "/login", req)
}

func (s *AuthTestSuite) TestRegisterSuccessful() {
	resp, err := s.registerUser("newuser", "password123")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Contains(resp.Header.Get("Set-Cookie"), "token=")
}

func (s *AuthTestSuite) TestRegisterValidationErrors() {
	testCases := []struct {
		name     string
		login    string
		password string
	}{
		{"empty_login", "", "pass"},
		{"short_login", "ab", "pass"},
		{"long_login", "aVeryLongLoginThatExceedsTheMaximumAllowedLengthOfFiftyCharactersWhichIsNotPermitted", "pass"},
		{"empty_password", "user", ""},
		{"short_password", "user", "12345"},
		{"long_password", "user", func() string {
			p := make([]byte, 101)
			for i := range p {
				p[i] = 'a'
			}
			return string(p)
		}()},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.registerUser(tc.login, tc.password)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func (s *AuthTestSuite) TestRegisterDuplicateLogin() {
	resp1, err := s.registerUser("duplicate", "password123")
	s.Require().NoError(err)
	defer resp1.Body.Close()
	s.Equal(http.StatusOK, resp1.StatusCode)

	resp2, err := s.registerUser("duplicate", "password123")
	s.Require().NoError(err)
	defer resp2.Body.Close()

	s.Equal(http.StatusConflict, resp2.StatusCode)
}

func (s *AuthTestSuite) TestLoginSuccessful() {
	respReg, err := s.registerUser("logintest", "password123")
	s.Require().NoError(err)
	defer respReg.Body.Close()
	s.Equal(http.StatusOK, respReg.StatusCode)

	respLogin, err := s.loginUser("logintest", "password123")
	s.Require().NoError(err)
	defer respLogin.Body.Close()

	s.Equal(http.StatusOK, respLogin.StatusCode)
	s.Contains(respLogin.Header.Get("Set-Cookie"), "token=")
}

func (s *AuthTestSuite) TestLoginValidationErrors() {
	testCases := []struct {
		name     string
		login    string
		password string
	}{
		{"empty_login", "", "pass"},
		{"empty_password", "user", ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.loginUser(tc.login, tc.password)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func (s *AuthTestSuite) TestLoginInvalidCredentials() {
	resp1, err := s.loginUser("unknown", "password")
	s.Require().NoError(err)
	defer resp1.Body.Close()
	s.Equal(http.StatusUnauthorized, resp1.StatusCode)

	respReg, err := s.registerUser("validuser", "realpass")
	s.Require().NoError(err)
	defer respReg.Body.Close()
	s.Equal(http.StatusOK, respReg.StatusCode)

	resp2, err := s.loginUser("validuser", "wrongpass")
	s.Require().NoError(err)
	defer resp2.Body.Close()
	s.Equal(http.StatusUnauthorized, resp2.StatusCode)
}

func (s *AuthTestSuite) TestRegisterJSONBindingError() {
	resp, err := s.performRequest("POST", "/register", `{ invalid json }`)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *AuthTestSuite) TestLoginJSONBindingError() {
	resp, err := s.performRequest("POST", "/login", `{ invalid json }`)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
