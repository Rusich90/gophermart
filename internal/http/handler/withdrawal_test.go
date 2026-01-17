package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	domainuser "github.com/Rusich90/gophermart.git/internal/domain/user"
	domainwithdrawal "github.com/Rusich90/gophermart.git/internal/domain/withdrawal"
	"github.com/Rusich90/gophermart.git/internal/http/authcontext"
	"github.com/Rusich90/gophermart.git/internal/http/handler"
	"github.com/Rusich90/gophermart.git/internal/repository"
	"github.com/Rusich90/gophermart.git/internal/service"
	"github.com/Rusich90/gophermart.git/internal/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type WithdrawalTestSuite struct {
	suite.Suite
	db                *pgxpool.Pool
	userRepo          domainuser.UserRepo
	orderRepo         domainorder.OrderRepo
	withdrawalRepo    domainwithdrawal.WithdrawalRepo
	httpSetup         *testutils.HTTPTestSetup
	jwtSecret         []byte
	logger            *zap.Logger
	withdrawalHandler *handler.WithdrawalHandler
	withdrawalService *service.WithdrawalService
	users             map[string]*uuid.UUID // Храним созданных пользователей для тестов
}

func (s *WithdrawalTestSuite) SetupSuite() {
	globalSuite, err := testutils.GetGlobalTestSuite()
	s.Require().NoError(err)

	s.db = globalSuite.DB
	s.userRepo = repository.NewUserRepo(s.db)
	s.orderRepo = repository.NewOrderRepo(s.db)
	s.withdrawalRepo = repository.NewWithdrawalRepo(s.db)
	s.jwtSecret = []byte("test_secret")
	s.logger = globalSuite.Logger
	s.users = make(map[string]*uuid.UUID)

	s.withdrawalService = service.NewWithdrawalService(s.withdrawalRepo)
	s.withdrawalHandler = handler.NewWithdrawalHandler(s.withdrawalService, s.logger)

	s.startServer()
}

func (s *WithdrawalTestSuite) startServer() {
	s.httpSetup = testutils.StartHTTPServer(func(r *gin.Engine) {
		// Добавляем middleware для установки userID в контекст для тестов
		r.Use(func(c *gin.Context) {
			// Проверяем, есть ли в заголовках специальный тестовый заголовок с userID
			if userIDStr := c.GetHeader("X-Test-User-ID"); userIDStr != "" {
				userID, err := uuid.Parse(userIDStr)
				if err == nil {
					authcontext.SetUserID(c, &userID)
				}
			}
			c.Next()
		})

		r.GET("/withdrawals", s.withdrawalHandler.GetAllByUserID)
	})
}

func (s *WithdrawalTestSuite) TearDownSuite() {
	if s.httpSetup != nil {
		s.httpSetup.Close()
	}
}

func (s *WithdrawalTestSuite) SetupTest() {
	s.truncateTables()
	s.createTestUsers()
}

func (s *WithdrawalTestSuite) truncateTables() {
	// Очищаем все таблицы перед каждым тестом
	tables := []string{"withdrawals", "orders", "users"}
	for _, table := range tables {
		_, err := s.db.Exec(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE;")
		s.Require().NoError(err, "Failed to truncate table "+table)
	}
}

func (s *WithdrawalTestSuite) createTestUsers() {
	s.users = make(map[string]*uuid.UUID)

	// Создаем тестовых пользователей
	testUsers := []struct {
		login    string
		password string
	}{
		{"user1", "password1"},
		{"user2", "password2"},
	}

	for _, user := range testUsers {
		userID := uuid.New()
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte(user.password), bcrypt.DefaultCost)
		domainUser := domainuser.User{
			ID:        userID,
			Login:     user.login,
			Password:  string(hashedPass),
			CreatedAt: time.Now(),
		}

		err := s.userRepo.Create(context.Background(), domainUser)
		s.Require().NoError(err)

		s.users[user.login] = &userID
	}
}

func (s *WithdrawalTestSuite) performRequest(method, url string, body interface{}, userID *uuid.UUID) (*http.Response, error) {
	var buf io.Reader
	if body != nil {
		// Если body уже строка, используем её как есть
		if str, ok := body.(string); ok {
			buf = bytes.NewBufferString(str)
		} else {
			// Иначе маршалим в JSON
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			buf = bytes.NewBuffer(b)
		}
	}

	fullURL := s.httpSetup.Server.URL + url
	req, err := http.NewRequest(method, fullURL, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Добавляем заголовок с userID для тестового middleware
	if userID != nil {
		req.Header.Set("X-Test-User-ID", userID.String())
	}

	client := &http.Client{}
	return client.Do(req)
}

func (s *WithdrawalTestSuite) createTestWithdrawal(userID *uuid.UUID, orderNum string, sum float64) {
	withdrawal := &domainwithdrawal.Withdrawal{
		OrderNum:  orderNum,
		UserID:    *userID,
		Sum:       sum,
		CreatedAt: time.Now(),
	}

	err := s.withdrawalRepo.Create(context.Background(), withdrawal)
	s.Require().NoError(err)
}

// =================== Tests ===================

func (s *WithdrawalTestSuite) TestGetAllByUserID_Successful() {
	userID := s.users["user1"]

	// Создаем тестовые данные - списания
	s.createTestWithdrawal(userID, "11111", 30.00)
	s.createTestWithdrawal(userID, "22222", 20.75)
	s.createTestWithdrawal(userID, "33333", 15.50)

	// Выполняем запрос
	resp, err := s.performRequest("GET", "/withdrawals", nil, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	// Проверяем содержимое ответа
	var withdrawalsResp []map[string]interface{}
	err = testutils.ParseJSONResponse(resp, &withdrawalsResp)
	s.Require().NoError(err)

	// Должно быть 3 записи, отсортированные по времени создания (новые первыми)
	s.Len(withdrawalsResp, 3)

	// Проверяем порядок (должен быть от новых к старым)
	s.Equal("33333", withdrawalsResp[0]["order"])
	s.Equal(15.50, withdrawalsResp[0]["sum"])
	s.Equal("22222", withdrawalsResp[1]["order"])
	s.Equal(20.75, withdrawalsResp[1]["sum"])
	s.Equal("11111", withdrawalsResp[2]["order"])
	s.Equal(30.00, withdrawalsResp[2]["sum"])
}

func (s *WithdrawalTestSuite) TestGetAllByUserID_NoData() {
	userID := s.users["user1"]

	// Выполняем запрос для пользователя без данных
	resp, err := s.performRequest("GET", "/withdrawals", nil, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusNoContent, resp.StatusCode)
}

func (s *WithdrawalTestSuite) TestGetAllByUserID_Unauthorized() {
	// Выполняем запрос без userID
	resp, err := s.performRequest("GET", "/withdrawals", nil, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

// =================== Entry Point ===================

func TestWithdrawalTestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawalTestSuite))
}
