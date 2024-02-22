package rabbit

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ilovepitsa/Customer/api/repo"
	// pb "github.com/ilovepitsa/Customer/protobuf"
	pb "github.com/ilovepitsa/protobufForTestCase"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type RabbitParameters struct {
	Login    string
	Password string
	Ip       string
	Port     string
}

type RabbitHandler struct {
	l               *log.Logger
	cr              *repo.CustomerRepository
	connection      *amqp.Connection
	channel         *amqp.Channel
	requestQueue    amqp.Queue
	responceQueue   amqp.Queue
	responceAccount amqp.Queue
}

func NewRabbitHandler(l *log.Logger, cr *repo.CustomerRepository) *RabbitHandler {
	return &RabbitHandler{l: l, cr: cr}
}

func (rb *RabbitHandler) Init(param RabbitParameters) error {
	var err error
	rb.connection, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", param.Login, param.Password, param.Ip, param.Port))
	if err != nil {
		// rb.l.Println(err)
		return err
	}

	rb.channel, err = rb.connection.Channel()
	if err != nil {
		// rb.l.Println(err)
		return err
	}

	err = rb.channel.ExchangeDeclare("customer", "topic", false, false, false, false, amqp.Table{})
	if err != nil {
		// rb.l.Println(err)
		return err
	}
	rb.requestQueue, err = rb.channel.QueueDeclare("customerRequest", false, false, false, false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}
	err = rb.channel.QueueBind(rb.requestQueue.Name, "request", "customer", false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	rb.responceQueue, err = rb.channel.QueueDeclare("customerReqsponce", false, false, false, false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	err = rb.channel.QueueBind(rb.responceQueue.Name, "responce", "customer", false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	rb.responceAccount, err = rb.channel.QueueDeclare("accountResponceC", false, false, false, false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}
	err = rb.channel.Qos(
		1,
		0,
		false,
	)

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	return nil
}

func (rb *RabbitHandler) Close() {
	rb.channel.Close()
	rb.connection.Close()
}

func (rb *RabbitHandler) Consume() {
	consumeRequestChan, err := rb.channel.Consume(rb.requestQueue.Name, "", true, false, false, false, amqp.Table{})
	consumeAccountResponce, err := rb.channel.Consume(rb.responceAccount.Name, "", true, false, false, false, amqp.Table{})
	if err != nil {
		rb.l.Println(err)
		return
	}
	var forever chan struct{}

	request := &pb.Request{}
	go func() {
		for d := range consumeRequestChan {
			err = proto.Unmarshal(d.Body, request)
			rb.l.Println("Recieve request ", request.String())
			switch request.Req.(type) {
			case *pb.Request_ReqAdd:
				rb.ParseRequest_Add(request.GetReqAdd())
			case *pb.Request_ReqGet:
				rb.ParseRequest_Get(request.GetReqGet())
			case *pb.Request_ReqGetAll:
				rb.ParseRequest_GetAll(request.GetReqGetAll())
			}
		}
	}()

	accountResponce := &pb.ResponceAccount{}
	go func() {
		for d := range consumeAccountResponce {
			err = proto.Unmarshal(d.Body, accountResponce)
			switch accountResponce.Resp.(type) {
			case *pb.ResponceAccount_RespActive:
				rb.parseAccountActive(accountResponce.GetRespActive())
			case *pb.ResponceAccount_RespFrozen:
				rb.parseAccountFrozen(accountResponce.GetRespFrozen())
			}
		}
	}()
	rb.l.Println("Waiting commands")
	<-forever
}

func (rb *RabbitHandler) parseAccountActive(resp *pb.ResponceActiveBalance) {
	invoices := resp.Invoices

	if len(invoices) < 1 {
		rb.l.Println("No active invoices customer id =", resp.RequestId)
		return
	}

	rb.l.Println("Active invoices  customer id =", resp.Invoices[0].CustomerId)
	for _, invoice := range invoices {
		rb.l.Println(invoice.Number)
	}
}

func (rb *RabbitHandler) parseAccountFrozen(resp *pb.ResponceFrozenBalance) {

	invoices := resp.Invoices

	if len(invoices) < 1 {
		rb.l.Println("No frozen invoices customer id =", resp.RequestId)
		return
	}

	rb.l.Println("Frozen invoices customer id =", resp.Invoices[0].CustomerId)
	for _, invoice := range invoices {
		rb.l.Println(invoice.Number)
	}
}

func (rb *RabbitHandler) ParseRequest_Add(req *pb.RequestAdd) {

	var newCust repo.Customer
	idToAnswer := req.RequestId
	newCust = repo.Customer{Name: req.Name}
	err := rb.cr.Add(newCust)
	isSucc := true
	if err != nil {
		isSucc = false
		rb.l.Println(err)
	}

	go rb.Responce_Add(idToAnswer, isSucc)

}

func (rb *RabbitHandler) Responce_Add(respId int32, isSucc bool) {
	var respAdd pb.ResponceAdd = pb.ResponceAdd{RequestId: respId, IsSuccess: isSucc}
	respProto := pb.Responce{Resp: &pb.Responce_RespAdd{RespAdd: &respAdd}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := proto.Marshal(&respProto)

	if err != nil {
		rb.l.Println(err)
		return
	}

	err = rb.channel.PublishWithContext(ctx,
		"customer",
		"responce",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        resp,
		})
	if err != nil {
		rb.l.Println(err)
		return
	}

}

func (rb *RabbitHandler) ParseRequest_Get(req *pb.RequestGet) {
	requestId, customerId := req.RequestId, req.CustomerId

	cust, err := rb.cr.Get(int(customerId))
	if err != nil {
		rb.l.Println(err)
		return
	}
	go rb.Responce_Get(requestId, cust)
}

func (rb *RabbitHandler) Responce_Get(respId int32, cust repo.Customer) {
	var respGet pb.ResponceGet = pb.ResponceGet{RequestId: respId, Name: cust.Name}
	respProto := pb.Responce{Resp: &pb.Responce_RespGet{RespGet: &respGet}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := proto.Marshal(&respProto)

	if err != nil {
		rb.l.Println(err)
		return
	}

	err = rb.channel.PublishWithContext(ctx,
		"customer",
		"responce",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        resp,
		})
	if err != nil {
		rb.l.Println(err)
		return
	}
}

func (rb *RabbitHandler) ParseRequest_GetAll(req *pb.RequestGetAll) {
	requestId := req.RequestId

	cust, err := rb.cr.GetAll()
	if err != nil {
		rb.l.Println(err)
		return
	}
	go rb.Responce_GetAll(requestId, cust)

}

func TOProtobuFCustomer(customers []repo.Customer) (result []*pb.Customer) {
	for _, v := range customers {
		result = append(result, &pb.Customer{Id: int32(v.Id), Name: v.Name})
	}
	return
}

func (rb *RabbitHandler) Responce_GetAll(respId int32, cust []repo.Customer) {

	pbCust := TOProtobuFCustomer(cust)
	var respGet pb.ResponceGetAll = pb.ResponceGetAll{Customers: pbCust}
	respProto := pb.Responce{Resp: &pb.Responce_RespGetAll{RespGetAll: &respGet}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := proto.Marshal(&respProto)

	if err != nil {
		rb.l.Println(err)
		return
	}

	err = rb.channel.PublishWithContext(ctx,
		"customer",
		"responce",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        resp,
		})
	if err != nil {
		rb.l.Println(err)
		return
	}
}

func (rb *RabbitHandler) RequestActive(id int) {
	req := pb.RequestActiveBalance{CustomerId: int32(id), RequestId: int32(id)}
	reqMess := pb.RequestAccount{Req: &pb.RequestAccount_ReqAct{ReqAct: &req}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reqByte, err := proto.Marshal(&reqMess)
	if err != nil {
		rb.l.Println(err)
		return
	}
	err = rb.channel.PublishWithContext(ctx,
		"account",
		"request",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        reqByte,
		})
	if err != nil {
		rb.l.Println(err)
		return
	}

}

func (rb *RabbitHandler) RequestFrozen(id int) {
	req := pb.RequestFrozenBalance{CustomerId: int32(id), RequestId: int32(id)}
	reqMess := pb.RequestAccount{Req: &pb.RequestAccount_ReqFroz{ReqFroz: &req}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reqByte, err := proto.Marshal(&reqMess)
	if err != nil {
		rb.l.Println(err)
		return
	}
	err = rb.channel.PublishWithContext(ctx,
		"account",
		"request",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        reqByte,
		})
	if err != nil {
		rb.l.Println(err)
		return
	}
}
