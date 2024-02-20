package main

import (
	"log"
	"net/http"
	"os"

	handlers "github.com/ilovepitsa/Customer/api/handlers/customerHandlers.go"
)

func main() {
	l := log.New(os.Stdout, "product-api", log.LstdFlags)
	custHandl := handlers.NewCustomerHandler(l)

	sm := http.NewServeMux()
	sm.Handle("/customers", custHandl)

	http.ListenAndServe(":8080", sm)
}
