package main

import "github.com/MaximkaSha/gophermart_loyalty/internal/service"

func main() {
	var app = service.NewService()
	app.StartService()
}
