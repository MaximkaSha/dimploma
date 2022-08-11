package service

import (
	"compress/flate"
	//"encoding/json"
	"log"
	"net/http"

	"github.com/MaximkaSha/gophermart_loyalty/internal/config"
	"github.com/MaximkaSha/gophermart_loyalty/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
)

type Service struct {
	cnfg config.Config
}

func NewService() *Service {
	return &Service{
		cnfg: config.NewConfig(),
	}
}

func (s *Service) StartService() {
	r := chi.NewRouter()
	logger := httplog.NewLogger("logger", httplog.Options{
		JSON:     false,
		LogLevel: "trace",
	})
	compressor := middleware.NewCompressor(flate.DefaultCompression)
	r.Use(compressor.Handler)
	r.Use(httplog.RequestLogger(logger))
	handler := handlers.NewHandlers(s.cnfg)
	go handler.Auth.SessionCleaner()

	//pub access
	r.Post("/api/user/register", handler.Register)
	r.Post("/api/user/login", handler.Login)

	//user only
	r.Group(func(r chi.Router) {
		r.Use(handler.CheckAuthMiddleWare)
		r.Post("/api/user/orders", handler.PostOrders)
		r.Get("/api/user/withdrawals", handler.GetWithdraws)
		//need to update balance operations
		r.Group(func(r chi.Router) {
			r.Use(handler.UpdateUserInfo)
			r.Post("/api/user/balance/withdraw", handler.PostWithdraw)
			r.Get("/api/user/orders", handler.GetOrders)
			r.Get("/api/user/balance", handler.GetBalance)
		})

	})
	log.Printf("Started service on %s", s.cnfg.Addr)
	if err := http.ListenAndServe(s.cnfg.Addr, r); err != nil {
		log.Printf("Server shutdown: %s", err.Error())
	}
}

/*
func (s Service) Test(w http.ResponseWriter, r *http.Request) {
	type JSONtest struct {
		Current   float32 `json:"current"`
		Withdrawn float32 `json:"withdrawn"`
	}
	tst := JSONtest{
		Current:   729.98,
		Withdrawn: 0,
	}
	JSNd, err := json.Marshal(tst)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(200)
	log.Println("HARDCODED")
	w.Write(JSNd)
}*/
