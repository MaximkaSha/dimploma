package accrual

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/MaximkaSha/gophermart_loyalty/internal/orders"
	"github.com/MaximkaSha/gophermart_loyalty/internal/utils"
)

type Accural struct {
	URL string
}

func NewAccural(URL string) Accural {
	if !utils.CheckURL(URL) {
		log.Println("Accural is not availble! Some functions are not availiable")
	}
	return Accural{
		URL: URL,
	}
}

func (a Accural) GetData(order orders.Order) (bool, orders.Order) {
	r, err := http.Get("http://" + a.URL + "/api/orders/" + fmt.Sprint(order.Number))
	if err != nil {
		log.Printf("Accural GET error: %s", err)
		return false, order
	}
	defer r.Body.Close()
	if r.StatusCode == 500 || r.StatusCode == 429 {
		log.Printf("Accural body parse error: %s", err)
		return false, order
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Accural body parse error: %s", err)
		return false, order
	}
	err = json.Unmarshal(body, &order)
	if err != nil {
		log.Printf("whoops: %s", err)
		return false, order
	}
	return true, order
}
