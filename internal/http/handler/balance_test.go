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
	"github.com/Rusich90/gophermart.git/internal/http/dto"
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

type BalanceTestSuite struct {
	suite.Suite
	db             *pgxpool.Pool
	userRepo       domainuser.UserRepo
	orderRepo      domainorder.OrderRepo
	withdrawalRepo domainwithdrawal.WithdrawalRepo
	httpSetup      *testutils.HTTPTestSetup
	jwtSecret      []byte
	logger         *zap.Logger
	balanceHandler *handler.BalanceHandler
	balanceService *service.BalanceService
	users          map[string]*uuid.UUID // Храним созданных пользователей для тестов
}

func (s *BalanceTestSuite) SetupSuite() {
	globalSuite, err := testutils.GetGlobalTestSuite()
	s.Require().NoError(err)
	
	s.db = globalSuite.DB
	s.userRepo = repository.NewUserRepo(s.db)
	s.orderRepo = repository.NewOrderRepo(s.db)
	s.withdrawalRepo = repository.NewWithdrawalRepo(s.db)
	s.jwtSecret = []byte("test_secret")
	s.logger = globalSuite.Logger
	s.users = make(map[string]*uuid.UUID)

	s.balanceService = service.NewBalanceService(s.orderRepo, s.withdrawalRepo)
	s.balanceHandler = handler.NewBalanceHandler(s.balanceService, s.logger)

	s.startServer()
}

func (s *BalanceTestSuite) startServer() {
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

		r.GET("/balance", s.balanceHandler.GetByUserID)
		r.POST("/balance/withdraw", s.balanceHandler.Withdraw)
	})
}

func (s *BalanceTestSuite) TearDownSuite() {
	if s.httpSetup != nil {
		s.httpSetup.Close()
	}
}

func (s *BalanceTestSuite) SetupTest() {
	s.truncateTables()
	s.createTestUsers()
}

func (s *BalanceTestSuite) truncateTables() {
	// Очищаем все таблицы перед каждым тестом
	tables := []string{"withdrawals", "orders", "users"}
	for _, table := range tables {
		_, err := s.db.Exec(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE;")
		s.Require().NoError(err, "Failed to truncate table "+table)
	}
}

func (s *BalanceTestSuite) createTestUsers() {
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

func (s *BalanceTestSuite) performRequest(method, url string, body interface{}, userID *uuid.UUID) (*http.Response, error) {
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

func (s *BalanceTestSuite) createTestOrder(userID *uuid.UUID, number string, status domainorder.OrderStatus, accrual float64) {
	order := &domainorder.Order{
		Number:    number,
		UserID:    *userID,
		Status:    status,
		Accrual:   accrual,
		CreatedAt: time.Now(),
	}
	
	err := s.orderRepo.Create(context.Background(), order)
	s.Require().NoError(err)
}

func (s *BalanceTestSuite) createTestWithdrawal(userID *uuid.UUID, orderNum string, sum float64) {
	withdrawal := &domainwithdrawal.Withdrawal{
		OrderNum:  orderNum,
		UserID:    *userID,
		Sum:       sum,
		CreatedAt: time.Now(),
	}
	
	err := s.withdrawalRepo.Create(context.Background(), withdrawal)
	s.Require().NoError(err)
}

func (s *BalanceTestSuite) TestGetByUserID_Successful() {
	userID := s.users["user1"]
	
	// Создаем тестовые данные - заказы с начислениями
	s.createTestOrder(userID, "12345", domainorder.PROCESSED, 100.50)
	s.createTestOrder(userID, "67890", domainorder.PROCESSED, 50.25)
	// Заказ в процессе обработки не должен учитываться
	s.createTestOrder(userID, "54321", domainorder.PROCESSING, 200.00)
	
	// Создаем тестовые данные - списания
	s.createTestWithdrawal(userID, "11111", 30.00)
	s.createTestWithdrawal(userID, "22222", 20.75)
	
	// Выполняем запрос
	resp, err := s.performRequest("GET", "/balance", nil, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusOK, resp.StatusCode)
	
	// Проверяем содержимое ответа
	var balanceResp dto.BalanceResponse
	err = testutils.ParseJSONResponse(resp, &balanceResp)
	s.Require().NoError(err)
	
	// Текущий баланс = (100.50 + 50.25) - (30.00 + 20.75) = 150.75 - 50.75 = 100.00
	s.Equal(100.00, balanceResp.Current)
	// Выведено = 30.00 + 20.75 = 50.75
	s.Equal(50.75, balanceResp.Withdrawn)
}

func (s *BalanceTestSuite) TestGetByUserID_NoData() {
	userID := s.users["user1"]
	
	// Выполняем запрос для пользователя без данных
	resp, err := s.performRequest("GET", "/balance", nil, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusOK, resp.StatusCode)
	
	// Проверяем содержимое ответа
	var balanceResp dto.BalanceResponse
	err = testutils.ParseJSONResponse(resp, &balanceResp)
	s.Require().NoError(err)
	
	// Балансы должны быть нулевыми
	s.Equal(0.00, balanceResp.Current)
	s.Equal(0.00, balanceResp.Withdrawn)
}

func (s *BalanceTestSuite) TestGetByUserID_Unauthorized() {
	// Выполняем запрос без userID
	resp, err := s.performRequest("GET", "/balance", nil, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *BalanceTestSuite) TestWithdraw_Successful() {
	userID := s.users["user1"]
	
	// Создаем тестовые данные - заказы с начислениями
	s.createTestOrder(userID, "12345678903", domainorder.PROCESSED, 100.00)
	
	// Подготавливаем запрос на вывод средств
	withdrawReq := dto.WithdrawRequest{
		Order: "11111111115", // Корректный номер по алгоритму Луна
		Sum:   50.00,
	}
	
	// Выполняем запрос
	resp, err := s.performRequest("POST", "/balance/withdraw", withdrawReq, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusOK, resp.StatusCode)
	
	// Проверяем, что запись создана в БД
	var count int
	err = s.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM withdrawals WHERE user_id = $1 AND order_num = $2 AND sum = $3",
	userID, "11111111115", 50.00).Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count)
}

func (s *BalanceTestSuite) TestWithdraw_InsufficientFunds() {
	userID := s.users["user1"]
	
	// Создаем тестовые данные - заказы с начислениями (меньше, чем хотим вывести)
	s.createTestOrder(userID, "12345678903", domainorder.PROCESSED, 30.00)
	
	// Подготавливаем запрос на вывод средств (больше, чем есть на счете)
	withdrawReq := dto.WithdrawRequest{
		Order: "11111111115", // Корректный номер по алгоритму Луна
		Sum:   50.00,
	}
	
	// Выполняем запрос
	resp, err := s.performRequest("POST", "/balance/withdraw", withdrawReq, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusPaymentRequired, resp.StatusCode)
	
	// Проверяем, что запись НЕ создана в БД
	var count int
	err = s.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM withdrawals WHERE user_id = $1 AND order_num = $2",
	userID, "11111111115").Scan(&count)
	s.Require().NoError(err)
	s.Equal(0, count)
}

func (s *BalanceTestSuite) TestWithdraw_InvalidOrderNumber() {
	userID := s.users["user1"]
	
	// Создаем тестовые данные - заказы с начислениями
	s.createTestOrder(userID, "12345678903", domainorder.PROCESSED, 100.00)
	
	// Подготавливаем запрос с некорректным номером заказа (не проходит проверку Луна)
	withdrawReq := dto.WithdrawRequest{
		Order: "12345678901", // Некорректный номер по алгоритму Луна
		Sum:   50.00,
	}
	
	// Выполняем запрос
	resp, err := s.performRequest("POST", "/balance/withdraw", withdrawReq, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)
	
	// Проверяем, что запись НЕ создана в БД
	var count int
	err = s.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM withdrawals WHERE user_id = $1 AND order_num = $2",
		userID, "12345678901").Scan(&count)
	s.Require().NoError(err)
	s.Equal(0, count)
}

func (s *BalanceTestSuite) TestWithdraw_OrderNumConflict() {
	userID := s.users["user1"]
	
	// Создаем тестовые данные - заказы с начислениями
	s.createTestOrder(userID, "12345678903", domainorder.PROCESSED, 100.00)
	
	// Создаем уже существующее списание с таким же номером заказа
	s.createTestWithdrawal(userID, "11111111115", 20.00)
	
	// Подготавливаем запрос с тем же номером заказа
	withdrawReq := dto.WithdrawRequest{
		Order: "11111111115", // Уже существует
		Sum:   50.00,
	}
	
	// Выполняем запрос
	resp, err := s.performRequest("POST", "/balance/withdraw", withdrawReq, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)
}

func (s *BalanceTestSuite) TestWithdraw_ValidationErrors() {
	userID := s.users["user1"]
	
	testCases := []struct {
		name          string
		request       dto.WithdrawRequest
		expectedCode  int
	}{
		{
			name: "empty_order",
			request: dto.WithdrawRequest{
				Order: "",
				Sum:   50.00,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "zero_sum",
			request: dto.WithdrawRequest{
				Order: "11111111115", // Корректный номер по алгоритму Луна
				Sum:   0,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "negative_sum",
			request: dto.WithdrawRequest{
				Order: "11111111115", // Корректный номер по алгоритму Луна
				Sum:   -10.00,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "too_small_sum",
			request: dto.WithdrawRequest{
				Order: "11111111115", // Корректный номер по алгоритму Луна
				Sum:   0.001,
			},
			expectedCode: http.StatusBadRequest,
		},
	}
	
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.performRequest("POST", "/balance/withdraw", tc.request, userID)
			s.Require().NoError(err)
			defer resp.Body.Close()
			
			s.Equal(tc.expectedCode, resp.StatusCode)
		})
	}
}

func (s *BalanceTestSuite) TestWithdraw_JSONBindingError() {
	userID := s.users["user1"]
	
	// Создаем запрос с некорректным JSON напрямую
	reqBody := `{ invalid json }`
	resp, err := s.performRequest("POST", "/balance/withdraw", reqBody, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *BalanceTestSuite) TestWithdraw_Unauthorized() {
	// Подготавливаем запрос на вывод средств
	withdrawReq := dto.WithdrawRequest{
		Order: "11111111115", // Корректный номер по алгоритму Луна
		Sum:   50.00,
	}
	
	// Выполняем запрос без userID
	resp, err := s.performRequest("POST", "/balance/withdraw", withdrawReq, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func TestBalanceTestSuite(t *testing.T) {
	suite.Run(t, new(BalanceTestSuite))
}