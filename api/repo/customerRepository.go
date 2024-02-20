package repo

import (
	"database/sql"
	"fmt"
	"log"
)

type Customer struct {
	Id   int
	Name string
}

type CustomerRepository struct {
	l  *log.Logger
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB, l *log.Logger) *CustomerRepository {
	return &CustomerRepository{l, db}
}

func (cr *CustomerRepository) Add(customer Customer) error {
	trans, err := cr.db.Begin()
	if err != nil {
		cr.l.Println(err)
		return err
	}
	stmt, err := trans.Exec(fmt.Sprintf("insert into customers (name) values ('%s')", customer.Name))
	if err != nil {
		cr.l.Println(err)
		return err
	}
	id, _ := stmt.LastInsertId()
	cr.l.Printf("Last inserted index: %v", id)
	trans.Commit()

	return nil
}

func (cr *CustomerRepository) GetAll() ([]Customer, error) {
	rows, err := cr.db.Query("Select * from customers;")
	if err != nil {
		cr.l.Print(err)
		return nil, err
	}
	defer rows.Close()
	var result []Customer
	for rows.Next() {
		cust := Customer{}
		err = rows.Scan(&cust.Id, &cust.Name)
		if err != nil {
			cr.l.Println(err)
			continue
		}
		result = append(result, cust)
	}
	return result, nil
}
