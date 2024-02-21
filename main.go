package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	handlers "github.com/ilovepitsa/Customer/api/handlers"
	"github.com/ilovepitsa/Customer/api/rabbit"
	repo "github.com/ilovepitsa/Customer/api/repo"
	_ "github.com/lib/pq"
)

func main() {
	l := log.New(os.Stdout, "Customers ", log.LstdFlags)

	connStr := "user=postgres password=123 dbname=TransactionSystem sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		l.Print(err)
		return
	}
	defer db.Close()

	custRepo := repo.NewCustomerRepository(db, l)

	custHandl := handlers.NewCustomerHandler(l, custRepo)
	rabbitHandler := rabbit.NewRabbitHandler(l, custRepo)

	err = rabbitHandler.Init(rabbit.RabbitParameters{Login: "customer", Password: "customer", Ip: "localhost", Port: "5672"})
	if err != nil {
		l.Println("Cant create rabbitHandler", err)
	}
	defer rabbitHandler.Close()

	go rabbitHandler.Consume()

	sm := http.NewServeMux()
	sm.Handle("/customers", custHandl)
	sm.Handle("/customer/", custHandl)

	l.Println("Starting .... ")
	http.ListenAndServe(":8080", sm)

}
