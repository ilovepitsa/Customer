package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	mux "github.com/gorilla/mux"
	handlers "github.com/ilovepitsa/Customer/api/handlers"
	"github.com/ilovepitsa/Customer/api/rabbit"
	repo "github.com/ilovepitsa/Customer/api/repo"
	_ "github.com/lib/pq"
)

func main() {
	l := log.New(os.Stdout, "Customers ", log.LstdFlags)
	l.SetFlags(log.LstdFlags | log.Lshortfile)

	connStr := "user=postgres password=123 dbname=TransactionSystem sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		l.Print(err)
		return
	}
	defer db.Close()

	custRepo := repo.NewCustomerRepository(db, l)

	rabbitHandler := rabbit.NewRabbitHandler(l, custRepo)

	err = rabbitHandler.Init(rabbit.RabbitParameters{Login: "customer", Password: "customer", Ip: "localhost", Port: "5672"})
	if err != nil {
		l.Println("Cant create rabbitHandler", err)
	}
	defer rabbitHandler.Close()

	go rabbitHandler.Consume()

	custHandl := handlers.NewCustomerHandler(l, custRepo, rabbitHandler)

	sm := mux.NewRouter()
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	postRouter := sm.Methods(http.MethodPost).Subrouter()
	// sm.Handle("/customers", custHandl)
	// sm.Handle("/customer/", custHandl)
	getRouter.HandleFunc("/customers", custHandl.GetAll)
	getRouter.HandleFunc("/customer/{id:[0-9]+}", custHandl.Get)
	getRouter.HandleFunc("/customer/{id:[0-9]+}/active", custHandl.GetActive)
	getRouter.HandleFunc("/customer/{id:[0-9]+}/frozen", custHandl.GetFrozen)
	postRouter.HandleFunc("/customers", custHandl.AddCustomer)

	l.Printf("Starting  on port %v.... \n", 8080)
	http.ListenAndServe(":8080", sm)

}
