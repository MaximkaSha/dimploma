package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/MaximkaSha/gophermart_loyalty/internal/accrual"
	"github.com/MaximkaSha/gophermart_loyalty/internal/config"
	"github.com/MaximkaSha/gophermart_loyalty/internal/models"
	"github.com/MaximkaSha/gophermart_loyalty/internal/orders"
	"github.com/MaximkaSha/gophermart_loyalty/internal/utils"
	_ "github.com/lib/pq"
)

type Storage struct {
	ConnectionString string
	Accural          accrual.Accural
	DB               *sql.DB
}

func NewStorage(cnfg config.Config) *Storage {
	s := new(Storage)
	s.ConnectionString = cnfg.DBaddr //"postgres://postgres:123456@127.0.0.1:5432/loyalty?sslmode=disable"
	err := s.InitDB()
	if err != nil {
		log.Panic("database error!")
	}
	s.Accural = accrual.NewAccural(cnfg.AccrualAddr)
	return s
}

func (s Storage) AddUser(m models.User) error {
	var query = `INSERT INTO users (username,password)
	VALUES ($1, $2)
	ON CONFLICT (username)
	DO UPDATE SET
	password = EXCLUDED.password`
	_, err := s.DB.Exec(query, m.Username, m.Password)
	if err != nil {
		log.Printf("Error %s when appending  data", err)
		return err
	}
	return err
}

func (s Storage) GetUser(m models.User) (models.User, error) {
	err := s.DB.QueryRow("SELECT username,password FROM users WHERE username = $1", m.Username).Scan(&m.Username, &m.Password)
	CheckError(err)
	return m, err
}

func (s *Storage) InitDB() error {
	psqlconn := s.ConnectionString
	var err error
	s.DB, err = sql.Open("postgres", psqlconn)
	CheckError(err)
	err = s.DB.Ping()
	CheckError(err)
	log.Println("DB Connected!")
	err = s.CreateDBIfNotExist()
	CheckError(err)
	err = s.CreateTableIfNotExist()
	CheckError(err)
	return err
}

func CheckError(err error) {
	if err != nil {
		log.Printf("Database error: %s", err)
	}
}

func (s Storage) CreateDBIfNotExist() error {
	var query = `SELECT 'CREATE DATABASE loyalty'
	WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'loyalty')`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err := s.DB.ExecContext(ctx, query)
	return err
}

func (s Storage) CreateTableIfNotExist() error {
	var query = `CREATE TABLE IF NOT EXISTS public.users
(
    id serial,
    username character varying(100) not null COLLATE pg_catalog."default",
    password character varying(100) not null COLLATE pg_catalog."default",
	current double precision,
    withdrawn double precision,
	primary key(username)
);
CREATE TABLE IF NOT EXISTS public.orders
(
    id serial,
    userid integer NOT NULL,
    ordernum bigint NOT NULL,
    accural double precision,
    status character varying(100) COLLATE pg_catalog."default" NOT NULL,
    upload_time character varying(100) COLLATE pg_catalog."default" NOT NULL,
	PRIMARY KEY (ordernum),
	CONSTRAINT unique_order UNIQUE (ordernum)
    
);
CREATE TABLE IF NOT EXISTS public.balance
(
    id serial ,
    userid bigint NOT NULL,
    current double precision,
    withdrawn double precision,
	primary key(userid)
);
CREATE TABLE IF NOT EXISTS public.history
(
    id serial NOT NULL,
    userid bigint NOT NULL,
    sum double precision NOT NULL,
    processed_at character varying(100) COLLATE pg_catalog."default" NOT NULL,
    "order" bigint NOT NULL,
    PRIMARY KEY ("order")
);
`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err := s.DB.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating  table", err)
		return err
	}

	return err
}

func (s Storage) AddOrder(order orders.Order, session models.Session) int {
	var query = `
	INSERT INTO orders (userid,ordernum,accural,status,upload_time)
	VALUES ((SELECT id from users where username = $1), $2, $3, $4 ,$5)`
	_, err := s.DB.Exec(query, session.Name, order.Number, order.Accural, order.Status, order.UploadedAt)
	if err != nil {
		//log.Printf("Error %s", err)
		var userName string
		query = `SELECT username from users where id = (SELECT userid from orders where ordernum = $1)`
		err := s.DB.QueryRow(query, order.Number).Scan(&userName)
		//log.Printf("session name: %s , db name: %s", session.Name, userName)
		if err != nil {
			log.Printf("Error %s", err)
			return 500
		}
		if session.Name == userName {
			log.Println("User already added this number")
			return 200
		} else {
			log.Println("Another user already added this number")
			return 409
		}
	}
	return 202
}

func (s Storage) GetAllOrders(session models.Session) (int, []orders.Order) {
	var query = `SELECT ordernum , status , accural , upload_time from orders where userid = (SELECT id from users where username = $1)`
	rows, err := s.DB.Query(query, session.Name)
	//	err = rows.Err()
	if err != nil {
		log.Printf("Error %s when getting all  data", err)
		return 204, []orders.Order{}
	}
	defer rows.Close()
	data := []orders.Order{}
	counter := 0
	for rows.Next() {
		model := orders.Order{}
		if err := rows.Scan(&model.Number, &model.Status, &model.Accural, &model.UploadedAt); err != nil {
			log.Println(err)
			return 500, nil
		}
		counter++
		data = append(data, model)
	}
	if counter == 0 {
		return 204, data
	}
	return 200, data
}

func (s Storage) GetAllOrdersToUpdate(session models.Session) (int, []orders.Order) {
	var query = `SELECT ordernum , status , accural , upload_time FROM public.orders WHERE status = 'NEW' OR status = 'PROCESSING' AND userid = (SELECT id from users where username = $1)`
	rows, err := s.DB.Query(query, session.Name)
	//	err = rows.Err()
	if err != nil {
		log.Printf("Error %s when getting all  data", err)
		return 204, []orders.Order{}
	}
	defer rows.Close()
	data := []orders.Order{}
	counter := 0
	for rows.Next() {
		model := orders.Order{}
		if err := rows.Scan(&model.Number, &model.Status, &model.Accural, &model.UploadedAt); err != nil {
			log.Println(err)
			return 500, nil
		}
		counter++
		data = append(data, model)
	}
	if counter == 0 {
		return 204, data
	}
	//log.Println(data)
	return 200, data
}

func (s Storage) GetBalance(session models.Session) (int, models.Balance) {
	var query = `SELECT current, withdrawn from balance WHERE userid = (SELECT id from users where username = $1)`
	data := models.Balance{}
	//log.Println(session)
	err := s.DB.QueryRow(query, session.Name).Scan(&data.Current, &data.Withdrawn)
	//log.Println(session)
	//log.Println(data)
	if err != nil {
		log.Printf("Error %s when getting balance data", err)
		return 204, models.Balance{}
	}
	return 200, data
}
func (s Storage) AddHistory(withdrawn models.Withdrawn, session models.Session) error {
	var query = `
	INSERT INTO history (userid,sum,processed_at,"order")
	VALUES ((SELECT id from users where username = $1), $2, $3, $4)`
	_, err := s.DB.Exec(query, session.Name, withdrawn.Sum, withdrawn.ProcessedAt, withdrawn.Order)
	if err != nil {
		log.Printf("Error adding history: %s", err)
		return errors.New("internal history error")
	}
	return nil
}
func (s Storage) UpdateBalance(balance models.Balance, session models.Session) error {
	var query = `INSERT INTO balance (userid,current,withdrawn) 
	VALUES ((SELECT id from users where username = $1),$2, $3)
	ON CONFLICT (userid)
	DO UPDATE SET
	current = EXCLUDED.current,
	withdrawn = EXCLUDED.withdrawn`
	_, err := s.DB.Exec(query, session.Name, balance.Current, balance.Withdrawn)
	if err != nil {
		log.Printf("Error updating balance: %s", err)
		return errors.New("internal update error")
	}
	return nil
}

func (s Storage) PostWithdraw(withdrawn models.Withdrawn, session models.Session) int {
	err, balance := s.GetBalance(session)
	if err != 200 {
		return 500
	}
	if balance.Current < withdrawn.Sum {
		return 402
	}
	balance.Withdrawn = balance.Withdrawn + withdrawn.Sum
	balance.Current = balance.Current - withdrawn.Sum
	e := s.UpdateBalance(balance, session)
	if e != nil {
		return 500
	}
	e = s.AddHistory(withdrawn, session)
	if e != nil {
		return 500
	}
	return 200
}

func (s Storage) GetHistory(session models.Session) (int, []models.Withdrawn) {
	var query = `SELECT sum , processed_at , "order"  from history where userid = (SELECT id from users where username = $1)`
	rows, err := s.DB.Query(query, session.Name)
	if err != nil {
		log.Printf("Error %s when getting all  data", err)
		return 500, []models.Withdrawn{}
	}
	defer rows.Close()
	data := []models.Withdrawn{}
	counter := 0
	for rows.Next() {
		model := models.Withdrawn{}
		if err := rows.Scan(&model.Sum, &model.ProcessedAt, &model.Order); err != nil {
			log.Println(err)
			return 500, nil
		}
		//log.Println(model)
		counter++
		data = append(data, model)
	}
	if counter == 0 {
		return 204, data
	}
	data = utils.SortSliceByRFC3339(data)
	return 200, data
}

func (s Storage) UpdateOrdersStatus(orders []orders.Order, session models.Session) {
	ret, balance := s.GetBalance(session)
	if ret != 200 {
		log.Printf("No balance data for user: %s", session.Name)
		//	return
	}
	for i := range orders {
		log.Printf("User before %s order #%s: %s\n Current balace: %f", session.Name, orders[i].Number, orders[i].Status, balance.Current)
		if upd, order := s.Accural.GetData(orders[i]); upd {
			orders[i] = order
			if order.Status == "PROCESSED" {
				balance.Current = balance.Current + order.Accural
			}
		}
		log.Printf("User after %s order #%s: %s\n Current balace: %f", session.Name, orders[i].Number, orders[i].Status, balance.Current)
	}
	err := s.BatchUpdateOrders(orders)
	if err != nil {
		log.Printf("Order status update error: %s", err)
	}
	//это наверное надо в пачку засунуть, а то получиться, что заказы обновились, а бабки не начислились
	// печальная будет ситуация :)
	err = s.UpdateBalance(balance, session)
	if err != nil {
		log.Println("Error updating balance on update.")
		return
	}
	log.Println(s.GetBalance(session))

}

func (s Storage) BatchUpdateOrders(orders []orders.Order) error {
	if len(orders) == 0 {
		return errors.New("empty batch")
	}
	var query = `UPDATE orders
	SET   accural=$1, status=$2 
	WHERE ordernum =$3;`
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, v := range orders {
		if _, err = stmt.Exec(v.Accural, v.Status, v.Number); err != nil {
			return err
		}
	}
	return tx.Commit()
}

/*
func (s Storage) TestAddOrder() {
	order := orders.Order{
		Number:     1111,
		Accural:    0.5,
		Status:     "NEW",
		UploadedAt: "123",
	}
	ses := models.Session{
		Name: "max2",
	}
	s.AddOrder(order, ses)
}

func (s Storage) TestAddOrder2() {
	order := orders.Order{
		Number:     2222283,
		Accural:    0.5,
		Status:     "NEW",
		UploadedAt: "123",
	}
	ses := models.Session{
		Name: "max",
	}
	s.AddOrder(order, ses)
}
*/
