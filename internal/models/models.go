package models

import (
	"time"

	"github.com/MaximkaSha/gophermart_loyalty/internal/orders"
)

/*
  {
      "current": 500.5,
      "withdrawn": 42
  }

*/

type User struct {
	Password string `json:"password"`
	Username string `json:"login"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawn struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func NewWithdrawn() Withdrawn {
	return Withdrawn{
		ProcessedAt: time.Now().Format(time.RFC3339),
	}
}

type Session struct {
	Token  string
	Expiry time.Time
	Name   string
}

type Storager interface {
	AddUser(User) error
	GetUser(User) (User, error)
	AddOrder(orders.Order, Session) int
	GetAllOrders(Session) (int, []orders.Order)
	GetBalance(Session) (int, Balance)
	PostWithdraw(Withdrawn, Session) int
	GetHistory(Session) (int, []Withdrawn)
	GetAllOrdersToUpdate(Session) (int, []orders.Order)
	UpdateOrdersStatus([]orders.Order, Session)
}

func (s Session) IsExpired() bool {
	return s.Expiry.Before(time.Now())
}
