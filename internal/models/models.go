package models

import (
	//"fmt"
	"strconv"
	"time"

	"github.com/MaximkaSha/gophermart_loyalty/internal/orders"
	//"github.com/shopspring/decimal"
)

/*
  {
      "current": 500.5,
      "withdrawn": 42
  }

*/

type Num float64
type CtxUserName string

type User struct {
	Password string `json:"password"`
	Username string `json:"login"`
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Withdrawn struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at,omitempty"`
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

/*
type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}
*/
func (n Num) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(n), 'f', 1, 32)), nil
}

func (c CtxUserName) String() string {
	return string(c)
}
