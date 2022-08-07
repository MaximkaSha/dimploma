package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr        string
	DBaddr      string
	AccrualAddr string
}

var (
	addr        string
	aBaddr      string
	accrualAddr string
)

func NewConfig() Config {
	cnfg := Config{}
	cnfg.Addr = ":80"
	cnfg.DBaddr = "postgres://postgres:123456@127.0.0.1:5432/loyalty?sslmode=disable"

	envData, ok := os.LookupEnv("RUN_ADDRESS")
	if ok {
		cnfg.Addr = envData
	}
	envData, ok = os.LookupEnv("DATABASE_URI")
	if ok {
		cnfg.DBaddr = envData
	}
	envData, ok = os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if ok {
		cnfg.AccrualAddr = envData
	}
	if flag.Lookup("a") == nil {
		flag.StringVar(&cnfg.Addr, "a", ":80", "service start address")
	}
	if flag.Lookup("d") == nil {
		flag.StringVar(&cnfg.DBaddr, "d", "postgres://postgres:123456@127.0.0.1:5432/loyalty?sslmode=disable", "DB URI address")
	}
	if flag.Lookup("r") == nil {
		flag.StringVar(&cnfg.AccrualAddr, "r", cnfg.AccrualAddr, " Accrual system address")
	}

	flag.Parse()
	return cnfg
}
