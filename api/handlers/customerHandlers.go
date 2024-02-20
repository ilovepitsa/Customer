package handlers

import (
	"fmt"
	"log"
	"net/http"

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
	customers, err := ch.cr.GetAll()
	if err != nil {
		fmt.Fprint(w, "Error access customers ", err)
		return
	}
	fmt.Fprintln(w, customers)

}
