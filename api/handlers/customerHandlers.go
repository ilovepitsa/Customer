package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ilovepitsa/Customer/api/rabbit"
	repo "github.com/ilovepitsa/Customer/api/repo"
)

type CustomerHandler struct {
	l      *log.Logger
	cr     *repo.CustomerRepository
	rabbit *rabbit.RabbitHandler
}

func NewCustomerHandler(l *log.Logger, cr *repo.CustomerRepository, rabbit *rabbit.RabbitHandler) *CustomerHandler {
	return &CustomerHandler{l: l, cr: cr, rabbit: rabbit}
}

func (ch *CustomerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if strings.EqualFold(r.URL.Path, "/customers") {
		ch.GetAll(w, r)
		return
	}

	if r.Method == http.MethodGet {
		fmt.Fprintln(w, r.URL.Path)
	}

	if r.Method == http.MethodPost {
		ch.AddCustomer(w, r)
	}
}

func (ch *CustomerHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	customers, err := ch.cr.GetAll()
	if err != nil {
		fmt.Fprint(w, "Error access customers ", err)
		return
	}
	fmt.Fprintln(w, customers)
}

func (ch *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprint(w, "WrongId ", err)
		return
	}

	customers, err := ch.cr.Get(id)
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

func (ch *CustomerHandler) GetActive(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprint(w, "WrongId ", err)
		return
	}

	ch.rabbit.RequestActive(id)
}

func (ch *CustomerHandler) GetFrozen(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Fprint(w, "WrongId ", err)
		return
	}

	ch.rabbit.RequestFrozen(id)
}
