package orders

import (
	"time"
	//	"golang.org/x/text/number"
)

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accural    float64 `json:"accural"`
	UploadedAt string  `json:"uploaded_at"`
}

func NewOrder(number string) Order {
	return Order{
		Number:     number,
		Status:     "NEW",
		UploadedAt: time.Now().Format(time.RFC3339),
	}
}
