package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	repo "github.com/ilovepitsa/Customer/api/repo"
)

type CustomerHandler struct {
	l  *log.Logger
	cr *repo.CustomerRepository
}

func NewCustomerHandler(l *log.Logger, cr *repo.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{l: l, cr: cr}
}

func (ch *CustomerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if strings.EqualFold(r.URL.Path, "/customers") {
		ch.GetAll(w)
		return
	}

	if r.Method == http.MethodGet {
		fmt.Fprintln(w, r.URL.Path)
	}

	if r.Method == http.MethodPost {
		ch.AddCustomer(w, r)
	}
}

func (ch *CustomerHandler) GetAll(w http.ResponseWriter) {
	customers, err := ch.cr.GetAll()
	if err != nil {
		fmt.Fprint(w, "Error access customers ", err)
		return
	}
	fmt.Fprintln(w, customers)
}

func (ch *CustomerHandler) AddCustomer(w http.ResponseWriter, r *http.Request) {
	ch.l.Println("Post method")

	cust := repo.Customer{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&cust)
	if err != nil {
		ch.l.Println("Error while pasring new customer")
		return
	}
	ch.cr.Add(cust)
}
