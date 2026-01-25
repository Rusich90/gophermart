package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"sync"
	"time"

	accrualclient "github.com/Rusich90/gophermart.git/internal/client/accrual"
	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	domainuser "github.com/Rusich90/gophermart.git/internal/domain/user"
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

// Мock для accrual клиента с возможностью контроля поведения и подсчета вызовов
type mockAccrualClient struct {
	mu        sync.Mutex
	response  *accrualclient.AccrualResponse
	err       error
	callCount int
}

func (m *mockAccrualClient) GetAccrualInfo(ctx context.Context, orderNumber string) (*accrualclient.AccrualResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.callCount++
	
	if m.response != nil {
		return m.response, m.err
	}
	
	return nil, m.err
}

func (m *mockAccrualClient) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

type OrderTestSuite struct {
	suite.Suite
	db           *pgxpool.Pool
	userRepo     domainuser.UserRepo
	orderRepo    domainorder.OrderRepo
	httpSetup    *testutils.HTTPTestSetup
	jwtSecret    []byte
	logger       *zap.Logger
	orderHandler *handler.OrderHandler
	orderService *service.OrderService
	users        map[string]*uuid.UUID // Храним созданных пользователей для тестов
}

func (s *OrderTestSuite) SetupSuite() {
	globalSuite, err := testutils.GetGlobalTestSuite()
	s.Require().NoError(err)

	s.db = globalSuite.DB
	s.userRepo = repository.NewUserRepo(s.db)
	s.orderRepo = repository.NewOrderRepo(s.db)
	s.jwtSecret = []byte("test_secret")
	s.logger = globalSuite.Logger
	s.users = make(map[string]*uuid.UUID)

	// Используем мок для accrual клиента
	mockClient := &mockAccrualClient{}
	s.orderService = service.NewOrderService(s.orderRepo, mockClient, s.logger, 5)
	s.orderHandler = handler.NewOrderHandler(s.orderService, s.logger)

	s.startServer()
}

func (s *OrderTestSuite) startServer() {
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

		r.GET("/orders", s.orderHandler.GetAllByUserID)
		r.POST("/orders", s.orderHandler.AddOrder)
	})
}

func (s *OrderTestSuite) TearDownSuite() {
	if s.httpSetup != nil {
		s.httpSetup.Close()
	}
}

func (s *OrderTestSuite) SetupTest() {
	s.truncateTables()
	s.createTestUsers()
}

func (s *OrderTestSuite) truncateTables() {
	// Очищаем все таблицы перед каждым тестом
	tables := []string{"orders", "users"}
	for _, table := range tables {
		_, err := s.db.Exec(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE;")
		s.Require().NoError(err, "Failed to truncate table "+table)
	}
}

func (s *OrderTestSuite) createTestUsers() {
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

func (s *OrderTestSuite) performRequest(method, url string, body interface{}, userID *uuid.UUID) (*http.Response, error) {
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
	req.Header.Set("Content-Type", "text/plain") // Для POST /orders

	// Добавляем заголовок с userID для тестового middleware
	if userID != nil {
		req.Header.Set("X-Test-User-ID", userID.String())
	}

	client := &http.Client{}
	return client.Do(req)
}

func (s *OrderTestSuite) createTestOrder(userID *uuid.UUID, number string, status domainorder.OrderStatus, accrual float64) {
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

func (s *OrderTestSuite) TestGetAllByUserID_Successful() {
	userID := s.users["user1"]

	// Создаем тестовые данные - заказы
	s.createTestOrder(userID, "12345678903", domainorder.NEW, 0)
	s.createTestOrder(userID, "09876543214", domainorder.PROCESSED, 100.50)
	s.createTestOrder(userID, "11111111115", domainorder.PROCESSING, 0)

	// Выполняем запрос
	resp, err := s.performRequest("GET", "/orders", nil, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	// Проверяем содержимое ответа
	var ordersResp []map[string]interface{}
	err = testutils.ParseJSONResponse(resp, &ordersResp)
	s.Require().NoError(err)

	// Должно быть 3 записи, отсортированные по времени создания (новые первыми)
	s.Len(ordersResp, 3)

	// Проверяем порядок (должен быть от новых к старым)
	s.Equal("11111111115", ordersResp[0]["number"])
	s.Equal("PROCESSING", ordersResp[0]["status"])
	s.Equal("09876543214", ordersResp[1]["number"])
	s.Equal("PROCESSED", ordersResp[1]["status"])
	s.Equal(100.50, ordersResp[1]["accrual"])
	s.Equal("12345678903", ordersResp[2]["number"])
	s.Equal("NEW", ordersResp[2]["status"])
}

func (s *OrderTestSuite) TestGetAllByUserID_NoData() {
	userID := s.users["user1"]

	// Выполняем запрос для пользователя без данных
	resp, err := s.performRequest("GET", "/orders", nil, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusNoContent, resp.StatusCode)
}

func (s *OrderTestSuite) TestGetAllByUserID_Unauthorized() {
	// Выполняем запрос без userID
	resp, err := s.performRequest("GET", "/orders", nil, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *OrderTestSuite) TestAddOrder_Successful() {
	userID := s.users["user1"]

	// Подготавливаем корректный номер заказа по алгоритму Луна
	orderNumber := "12345678903" // Корректный номер

	// Выполняем запрос
	resp, err := s.performRequest("POST", "/orders", orderNumber, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusAccepted, resp.StatusCode)

	// Проверяем, что запись создана в БД
	var count int
	err = s.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM orders WHERE user_id = $1 AND number = $2",
		userID, orderNumber).Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count)
}

func (s *OrderTestSuite) TestAddOrder_AlreadyUploadedBySameUser() {
	userID := s.users["user1"]

	// Создаем заказ в БД
	orderNumber := "12345678903"
	s.createTestOrder(userID, orderNumber, domainorder.NEW, 0)

	// Пытаемся добавить тот же заказ
	resp, err := s.performRequest("POST", "/orders", orderNumber, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *OrderTestSuite) TestAddOrder_OwnedByOtherUser() {
	user1ID := s.users["user1"]
	user2ID := s.users["user2"]

	// Создаем заказ для другого пользователя
	orderNumber := "12345678903"
	s.createTestOrder(user2ID, orderNumber, domainorder.NEW, 0)

	// Пытаемся добавить чужой заказ
	resp, err := s.performRequest("POST", "/orders", orderNumber, user1ID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusConflict, resp.StatusCode)
}

func (s *OrderTestSuite) TestAddOrder_InvalidOrderNumber() {
	userID := s.users["user1"]

	// Подготавливаем некорректный номер заказа (не проходит проверку Луна)
	orderNumber := "12345678901" // Некорректный номер

	// Выполняем запрос
	resp, err := s.performRequest("POST", "/orders", orderNumber, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)
}

func (s *OrderTestSuite) TestAddOrder_EmptyBody() {
	userID := s.users["user1"]

	// Выполняем запрос с пустым телом
	resp, err := s.performRequest("POST", "/orders", "", userID)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *OrderTestSuite) TestAddOrder_Unauthorized() {
	// Подготавливаем корректный номер заказа по алгоритму Луна
	orderNumber := "12345678903" // Корректный номер

	// Выполняем запрос без userID
	resp, err := s.performRequest("POST", "/orders", orderNumber, nil)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *OrderTestSuite) TestAddOrder_CallsProcessAccrualAsync() {
	// Создаем мок клиента с подсчетом вызовов
	mockClient := &mockAccrualClient{
		response: &accrualclient.AccrualResponse{
			Order:   "12345678903",
			Status:  accrualclient.StatusProcessed,
			Accrual: nil,
		},
	}
	
	// Переопределяем orderService с новым моком
	s.orderService = service.NewOrderService(s.orderRepo, mockClient, s.logger, 10*time.Millisecond)
	s.orderHandler = handler.NewOrderHandler(s.orderService, s.logger)
	
	// Перезапускаем сервер с новым хендлером
	s.startServer()
	
	userID := s.users["user1"]
	orderNumber := "12345678903"
	
	// Выполняем запрос
	resp, err := s.performRequest("POST", "/orders", orderNumber, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusAccepted, resp.StatusCode)
	
	// Ждем немного, чтобы дать время горутине выполниться
	time.Sleep(100 * time.Millisecond)
	
	// Проверяем, что метод GetAccrualInfo был вызван хотя бы один раз
	s.True(mockClient.GetCallCount() > 0, "Expected GetAccrualInfo to be called at least once")
}

// Тест для проверки обработки заказа со статусом INVALID
func (s *OrderTestSuite) TestAddOrder_ProcessInvalidOrder() {
	// Создаем мок клиента, который будет возвращать статус INVALID
	mockClient := &mockAccrualClient{
		response: &accrualclient.AccrualResponse{
			Order:  "12345678903",
			Status: accrualclient.StatusInvalid,
		},
	}
	
	// Переопределяем orderService с новым моком и маленьким интервалом
	s.orderService = service.NewOrderService(s.orderRepo, mockClient, s.logger, 10*time.Millisecond)
	s.orderHandler = handler.NewOrderHandler(s.orderService, s.logger)
	
	// Перезапускаем сервер с новым хендлером
	s.startServer()
	
	userID := s.users["user1"]
	orderNumber := "12345678903"
	
	// Выполняем запрос
	resp, err := s.performRequest("POST", "/orders", orderNumber, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusAccepted, resp.StatusCode)
	
	// Ждем немного, чтобы дать время горутине выполниться
	time.Sleep(100 * time.Millisecond)
	
	// Проверяем, что метод GetAccrualInfo был вызван
	s.True(mockClient.GetCallCount() > 0, "Expected GetAccrualInfo to be called")
	
	// Проверяем, что статус заказа в БД стал INVALID
	order, err := s.orderRepo.GetByNumber(context.Background(), orderNumber)
	s.Require().NoError(err)
	s.Equal(domainorder.INVALID, order.Status)
}

// Тест для проверки обработки заказа со статусом PROCESSED и ненулевым accrual
func (s *OrderTestSuite) TestAddOrder_ProcessValidOrderWithAccrual() {
	// Создаем мок клиента, который будет возвращать статус PROCESSED с ненулевым accrual
	accrualValue := 100.50
	mockClient := &mockAccrualClient{
		response: &accrualclient.AccrualResponse{
			Order:   "12345678903",
			Status:  accrualclient.StatusProcessed,
			Accrual: &accrualValue,
		},
	}
	
	// Переопределяем orderService с новым моком и маленьким интервалом
	s.orderService = service.NewOrderService(s.orderRepo, mockClient, s.logger, 10*time.Millisecond)
	s.orderHandler = handler.NewOrderHandler(s.orderService, s.logger)
	
	// Перезапускаем сервер с новым хендлером
	s.startServer()
	
	userID := s.users["user1"]
	orderNumber := "12345678903"
	
	// Выполняем запрос
	resp, err := s.performRequest("POST", "/orders", orderNumber, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusAccepted, resp.StatusCode)
	
	// Ждем немного, чтобы дать время горутине выполниться
	time.Sleep(100 * time.Millisecond)
	
	// Проверяем, что метод GetAccrualInfo был вызван
	s.True(mockClient.GetCallCount() > 0, "Expected GetAccrualInfo to be called")
	
	// Проверяем, что статус заказа в БД стал PROCESSED и accrual обновился
	order, err := s.orderRepo.GetByNumber(context.Background(), orderNumber)
	s.Require().NoError(err)
	s.Equal(domainorder.PROCESSED, order.Status)
	s.Equal(accrualValue, order.Accrual)
}

// Мock для accrual клиента с возможностью возвращать ошибки
type errorAccrualClient struct {
	mu        sync.Mutex
	err       error
	callCount int
	maxCalls  int
}

func (e *errorAccrualClient) GetAccrualInfo(ctx context.Context, orderNumber string) (*accrualclient.AccrualResponse, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.callCount++
	
	// Если задано максимальное количество вызовов, то возвращаем ошибку до тех пор,
	// пока не достигнем этого количества
	if e.maxCalls > 0 && e.callCount <= e.maxCalls {
		return nil, e.err
	}
	
	// После достижения нужного количества вызовов возвращаем успешный результат
	return &accrualclient.AccrualResponse{
		Order:   orderNumber,
		Status:  accrualclient.StatusProcessed,
		Accrual: nil,
	}, nil
}

func (e *errorAccrualClient) GetCallCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.callCount
}

// Тест для проверки механизма повторных попыток при сетевых ошибках
func (s *OrderTestSuite) TestAddOrder_RetryOnNetworkErrors() {
	// Создаем мок клиента, который будет возвращать ошибку 3 раза, а затем успешный результат
	errorClient := &errorAccrualClient{
		err:      accrualclient.ErrInternalServer,
		maxCalls: 3,
	}
	
	// Переопределяем orderService с новым моком и маленьким интервалом
	s.orderService = service.NewOrderService(s.orderRepo, errorClient, s.logger, 10*time.Millisecond)
	s.orderHandler = handler.NewOrderHandler(s.orderService, s.logger)
	
	// Перезапускаем сервер с новым хендлером
	s.startServer()
	
	userID := s.users["user1"]
	orderNumber := "12345678903"
	
	// Выполняем запрос
	resp, err := s.performRequest("POST", "/orders", orderNumber, userID)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Equal(http.StatusAccepted, resp.StatusCode)
	
	// Ждем достаточно долго, чтобы произошли все повторные попытки
	time.Sleep(500 * time.Millisecond)
	
	// Проверяем, что метод GetAccrualInfo был вызван 4 раза (3 раза с ошибкой + 1 успешный)
	s.Equal(4, errorClient.GetCallCount(), "Expected GetAccrualInfo to be called 4 times")
	
	// Проверяем, что статус заказа в БД стал PROCESSED
	order, err := s.orderRepo.GetByNumber(context.Background(), orderNumber)
	s.Require().NoError(err)
	s.Equal(domainorder.PROCESSED, order.Status)
}

func TestOrderTestSuite(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}
