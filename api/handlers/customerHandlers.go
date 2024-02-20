package handlers

import (
	"fmt"
	"log"
	"net/http"
)

type CustomerHandler struct {
	l *log.Logger
}

func NewCustomerHandler(l *log.Logger) *CustomerHandler {
	return &CustomerHandler{l: l}
}

func (ch *CustomerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "helloFromCustomer")
	ch.l.Println("Get all Customers")
}
