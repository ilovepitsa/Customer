package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	handlers "github.com/ilovepitsa/Customer/api/handlers"
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

	// custRepo.Add(repo.Customer{Name: "Nikita"})
	// custRepo.Add(repo.Customer{Name: "Dasha"})
	// custRepo.Add(repo.Customer{Name: "Gosha"})
	// custRepo.Add(repo.Customer{Name: "Sasha"})

	sm := http.NewServeMux()
	sm.Handle("/customers", custHandl)

	http.ListenAndServe(":8080", sm)
}
