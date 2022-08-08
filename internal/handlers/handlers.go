package handlers

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/MaximkaSha/gophermart_loyalty/internal/auth"
	"github.com/MaximkaSha/gophermart_loyalty/internal/config"
	"github.com/MaximkaSha/gophermart_loyalty/internal/models"
	"github.com/MaximkaSha/gophermart_loyalty/internal/orders"
	"github.com/MaximkaSha/gophermart_loyalty/internal/storage"

	"github.com/google/uuid"
	"github.com/theplant/luhn"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	Store models.Storager
	Auth  auth.Auth
}

func NewHandlers(cnfg config.Config) *Handlers {
	return &Handlers{
		Store: storage.NewStorage(cnfg),
		Auth:  auth.NewAuth(),
	}
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var NewUser = &models.User{}
	err := json.NewDecoder(r.Body).Decode(NewUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(NewUser.Password), 14)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	NewUser.Password = string(hash)
	err = h.Store.AddUser(*NewUser)
	if err != nil {
		w.WriteHeader(409)
		return
	}
	c := models.Session{}
	c.Token = uuid.NewString()
	c.Expiry = time.Now().Add(120 * time.Second)
	c.Name = NewUser.Username
	h.Auth.AddSession(c)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   c.Token,
		Expires: c.Expiry,
	})

	w.WriteHeader(http.StatusOK)
	return

}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userDB, err := h.Store.GetUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(user.Password))
	if err != nil {
		w.WriteHeader(401)
		return
	}
	c := models.Session{}
	c.Token = uuid.NewString()
	c.Expiry = time.Now().Add(120 * time.Second)
	c.Name = user.Username
	h.Auth.AddSession(c)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   c.Token,
		Expires: c.Expiry,
		Path:    "/",
	})

	w.WriteHeader(http.StatusOK)
	return

}

func (h *Handlers) CheckAuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		token := c.Value
		session, err := h.Auth.GetSessionByUUID(token)
		if err != nil {
			log.Println("cookie not found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if session.IsExpired() {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) PostOrders(w http.ResponseWriter, r *http.Request) {
	var orderBuf []byte
	orderBuf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()
	//log.Printf("body: %s", orderBuf)
	orderNum, err := strconv.Atoi(string(orderBuf))
	if err != nil {
		log.Println(err)
		w.WriteHeader(422)
		return
	}
	if !luhn.Valid(int(orderNum)) {
		w.WriteHeader(422)
		return
	}
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	token := c.Value
	session, err := h.Auth.GetSessionByUUID(token)
	order := orders.NewOrder(string(orderBuf))
	w.WriteHeader(h.Store.AddOrder(order, session))
	return
}

func (h *Handlers) GetOrders(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	token := c.Value
	session, err := h.Auth.GetSessionByUUID(token)
	orders := []orders.Order{}
	ret, orders := h.Store.GetAllOrders(session)
	//log.Println(orders)
	JSONdata, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(ret)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ret)
	//log.Println(hex.EncodeToString(JSONdata))
	w.Write(JSONdata)
	return
}

func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	token := c.Value
	session, err := h.Auth.GetSessionByUUID(token)
	balance := models.Balance{}
	ret, balance := h.Store.GetBalance(session)
	log.Printf("Get balance data:%f,%f", balance.Current, balance.Withdrawn)
	JSONdata, err := json.Marshal(balance)
	if err != nil {
		log.Printf("Balance marshal error %s", err)
		w.WriteHeader(ret)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ret)
	log.Println("dump balance data:")
	log.Println(hex.EncodeToString(JSONdata))
	w.Write(JSONdata)
	return
}

func (h *Handlers) PostWithdraw(w http.ResponseWriter, r *http.Request) {
	withdrawn := models.NewWithdrawn()
	err := json.NewDecoder(r.Body).Decode(&withdrawn)
	if err != nil {
		log.Println("cant unmarshal")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderNum, err := strconv.Atoi(withdrawn.Order)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	if !luhn.Valid(int(orderNum)) {
		w.WriteHeader(422)
		return
	}
	c, err := r.Cookie("session_token")
	token := c.Value
	session, err := h.Auth.GetSessionByUUID(token)
	ret := h.Store.PostWithdraw(withdrawn, session)
	w.WriteHeader(ret)
	return
}

func (h *Handlers) GetWithdraws(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	token := c.Value
	session, err := h.Auth.GetSessionByUUID(token)
	ret, history := h.Store.GetHistory(session)
	//utils.SortSliceByRFC3339(history)
	JSONdata, err := json.Marshal(history)
	if err != nil {
		w.WriteHeader(ret)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ret)
	w.Write(JSONdata)
	return
}

func (h *Handlers) UpdateUserInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		token := c.Value
		session, err := h.Auth.GetSessionByUUID(token)
		_, orders := h.Store.GetAllOrdersToUpdate(session)
		//log.Println(orders)
		if len(orders) == 0 {
			next.ServeHTTP(w, r)
		}
		h.Store.UpdateOrdersStatus(orders, session)
		next.ServeHTTP(w, r)
	})
}
