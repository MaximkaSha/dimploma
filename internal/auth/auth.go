package auth

import (
	"errors"
	"time"

	"github.com/MaximkaSha/gophermart_loyalty/internal/models"
)

type Auth struct {
	sessions map[string]sessionCache
}

type sessionCache struct {
	username string
	expiry   time.Time
}

func NewAuth() Auth {
	return Auth{
		sessions: make(map[string]sessionCache),
	}
}
func (a Auth) AddSession(ses models.Session) error {
	a.sessions[ses.Token] = sessionCache{
		username: ses.Name,
		expiry:   ses.Expiry,
	}
	return nil
}

func (a Auth) GetSession(ses models.Session) (models.Session, error) {
	if val, ok := a.sessions[ses.Token]; ok {
		data := models.Session{
			Token:  ses.Token,
			Name:   val.username,
			Expiry: val.expiry,
		}
		return data, nil
	}
	return models.Session{}, errors.New("no data")
}

func (a Auth) GetSessionByUUID(token string) (models.Session, error) {
	if val, ok := a.sessions[token]; ok {
		data := models.Session{
			Token:  token,
			Name:   val.username,
			Expiry: val.expiry,
		}
		return data, nil
	}
	return models.Session{}, errors.New("no data")
}

func (a Auth) SessionCleaner() {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				for i := range a.sessions {
					if a.sessions[i].expiry.Before(time.Now()) {
						delete(a.sessions, i)
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
