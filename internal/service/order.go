package service

import (
	"context"
	"fmt"
	"time"

	accrualclient "github.com/Rusich90/gophermart.git/internal/client/accrual"
	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	"github.com/Rusich90/gophermart.git/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderService struct {
	orderRepo      domainorder.OrderRepo
	accrualClient  accrualclient.AccrualInterface
	logger         *zap.Logger
	tickerDuration time.Duration
}

func NewOrderService(orderRepo domainorder.OrderRepo, accrualClient accrualclient.AccrualInterface, logger *zap.Logger, tickerDuration time.Duration) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		accrualClient:  accrualClient,
		logger:         logger,
		tickerDuration: tickerDuration,
	}
}

func (s *OrderService) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainorder.Order, error) {
	orders, err := s.orderRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.GetAllByUserID: %w", err)
	}

	return orders, nil
}

func (s *OrderService) AddOrder(ctx context.Context, userID *uuid.UUID, number string) error {
	if !utils.IsValidLuhn(number) {
		return domainorder.ErrInvalidOrderNumber
	}

	existOrder, err := s.orderRepo.GetByNumber(ctx, number)
	if err != nil {
		if !domainorder.IsErrOrderNotFound(err) {
			return fmt.Errorf("failed to check order existence: %w", err)
		}
	} else {
		if existOrder.UserID != *userID {
			return domainorder.ErrOrderOwnedByOtherUser
		}
		return domainorder.ErrOrderAlreadyUploaded
	}

	order := &domainorder.Order{
		Number:    number,
		UserID:    *userID,
		Status:    domainorder.NEW,
		Accrual:   0,
		CreatedAt: time.Now(),
	}

	err = s.orderRepo.Create(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	go s.processAccrualAsync(number)

	return nil
}

func (s *OrderService) processAccrualAsync(orderNumber string) {
	ctx := context.Background()
	ticker := time.NewTicker(s.tickerDuration)
	defer ticker.Stop()

	const maxRetries = 5
	retries := 0

	for {
		select {
		case <-ticker.C:
			result, err := s.accrualClient.GetAccrualInfo(ctx, orderNumber)
			if err != nil {
				switch {
				case accrualclient.IsErrTooManyRequests(err):
					retries++
					if retries >= maxRetries {
						s.logger.Error("Max retries exceeded due to too many requests",
							zap.String("order_number", orderNumber))
						return
					}
					time.Sleep(s.tickerDuration * 4)
					continue

				case accrualclient.IsErrInternalServer(err):
					retries++
					if retries >= maxRetries {
						s.logger.Error("Max retries exceeded due to internal server error",
							zap.String("order_number", orderNumber),
							zap.Error(err))
						return
					}
					time.Sleep(s.tickerDuration)
					continue

				case accrualclient.IsErrOrderNotRegistered(err):
					s.logger.Warn("Order is not registered yet", zap.String("order_number", orderNumber))
					s.markOrderAsInvalid(ctx, orderNumber)
					return

				default:
					s.logger.Error("Fatal error fetching accrual data",
						zap.String("order_number", orderNumber),
						zap.Error(err))
					return
				}
			}

			retries = 0

			switch result.Status {
			case accrualclient.StatusInvalid:
				s.markOrderAsInvalid(ctx, orderNumber)
				return

			case accrualclient.StatusProcessed:
				accrual := 0.0
				if result.Accrual != nil {
					accrual = *result.Accrual
				}
				s.markOrderAsProcessed(ctx, orderNumber, accrual)
				return

			default:
				continue
			}

		case <-ctx.Done():
			s.logger.Info("Context cancelled, stopping accrual polling",
				zap.String("order_number", orderNumber))
			return
		}
	}
}

func (s *OrderService) markOrderAsInvalid(ctx context.Context, orderNumber string) {
	order := &domainorder.Order{
		Number:  orderNumber,
		Status:  domainorder.INVALID,
		Accrual: 0.0,
	}
	err := s.orderRepo.Update(ctx, order)
	if err != nil {
		s.logger.Error("Failed to update order to INVALID",
			zap.String("order_number", orderNumber),
			zap.Error(err))
	}
}

func (s *OrderService) markOrderAsProcessed(ctx context.Context, orderNumber string, accrual float64) {
	order := &domainorder.Order{
		Number:  orderNumber,
		Status:  domainorder.PROCESSED,
		Accrual: accrual,
	}
	err := s.orderRepo.Update(ctx, order)
	if err != nil {
		s.logger.Error("Failed to update order to PROCESSED",
			zap.String("order_number", orderNumber),
			zap.Error(err))
	}
}
