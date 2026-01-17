package testutils

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	globalDBSetup *DBSetup
	once          sync.Once
	setupErr      error
	mutex         sync.Mutex
)

// GlobalTestSuite представляет собой глобальный тестовый набор, который инициализируется один раз для всех тестов
type GlobalTestSuite struct {
	DB     *pgxpool.Pool
	Logger *zap.Logger
}

// GetGlobalTestSuite возвращает глобальный тестовый набор, инициализируя его при первом вызове
func GetGlobalTestSuite() (*GlobalTestSuite, error) {
	once.Do(func() {
		globalDBSetup, setupErr = SetupTestDB()
	})

	if setupErr != nil {
		return nil, fmt.Errorf("failed to setup global test database: %w", setupErr)
	}

	return &GlobalTestSuite{
		DB:     globalDBSetup.DB,
		Logger: zap.NewNop(),
	}, nil
}

// CleanupGlobalTestSuite очищает глобальный тестовый набор
func CleanupGlobalTestSuite() {
	mutex.Lock()
	defer mutex.Unlock()
	
	if globalDBSetup != nil {
		CloseDBSetup(globalDBSetup)
		globalDBSetup = nil
		once = sync.Once{} // Сброс once для возможности повторной инициализации в других тестах
	}
}

// TruncateAllTables очищает все таблицы в базе данных
func (g *GlobalTestSuite) TruncateAllTables() error {
	mutex.Lock()
	defer mutex.Unlock()
	
	// Проверяем, что пул соединений еще не закрыт
	if g.DB == nil {
		return fmt.Errorf("database pool is closed")
	}
	
	tables := []string{"withdrawals", "orders", "users"}
	for _, table := range tables {
		_, err := g.DB.Exec(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE;")
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}
	return nil
}

// SetupGlobalTestSuite инициализирует глобальный тестовый набор для использования в TestMain
func SetupGlobalTestSuite(m *testing.M) {
	suite, err := GetGlobalTestSuite()
	if err != nil {
		fmt.Printf("Failed to setup global test suite: %v\n", err)
		os.Exit(1)
	}

	// Очищаем таблицы перед запуском тестов
	if err := suite.TruncateAllTables(); err != nil {
		fmt.Printf("Failed to truncate tables: %v\n", err)
		os.Exit(1)
	}

	// Запуск тестов
	code := m.Run()

	// Очистка после завершения всех тестов
	CleanupGlobalTestSuite()

	// Выход с кодом завершения тестов
	os.Exit(code)
}