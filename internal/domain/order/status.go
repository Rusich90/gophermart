package order

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type OrderStatus string

const (
	NEW        OrderStatus = "NEW"
	PROCESSING OrderStatus = "PROCESSING"
	INVALID    OrderStatus = "INVALID"
	PROCESSED  OrderStatus = "PROCESSED"
)

func (s OrderStatus) String() string {
	return string(s)
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case NEW, PROCESSING, INVALID, PROCESSED:
		return true
	default:
		return false
	}
}

func (s OrderStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	if sv, ok := value.(string); ok {
		status := OrderStatus(sv)

		if !status.IsValid() {
			return fmt.Errorf("invalid order status: %q", sv)
		}

		*s = status
		return nil
	}

	return fmt.Errorf("cannot scan %T into OrderStatus", value)
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *OrderStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	status := OrderStatus(str)
	if !status.IsValid() {
		return fmt.Errorf("invalid order status: %s", str)
	}
	*s = status
	return nil
}
