/customers - get all customers
/customer/id - get customer with this id
/customer/id/active - get active customer's invoice with this id
/customer/id/frozen - get frozen customer's invoice with this id

post --> /customers - add customer

SQL:
table name: customer
sql script:
CREATE TABLE IF NOT EXISTS public.customer
(
    "customerId" integer NOT NULL DEFAULT nextval('"customer_customerId_seq"'::regclass),
    name text COLLATE pg_catalog."default",
    CONSTRAINT customer_pkey PRIMARY KEY ("customerId")
)

rabbit:
    host: /
    user: customer