package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	proto "replication/proto"
	"strconv"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedBankServer
	name    string
	port    int
	balance int32
	owner   int32
}

var port = flag.Int("port", 0, "server port number") // create the port that recieves the port that the client wants to access to

func main() {
	f := setLog()
	defer f.Close()

	flag.Parse()

	server := &Server{
		name:  "serverName",
		port:  *port,
		owner: -1,
	}

	go startServer(server)

	for {

	}
}

func startServer(server *Server) {
	grpcServer := grpc.NewServer()                                           // create a new grpc server
	listen, err := net.Listen("tcp", "localhost:"+strconv.Itoa(server.port)) // creates the listener

	if err != nil {
		log.Fatalln("Could not start listener")
	}

	log.Printf("Server started at port %v", server.port)

	proto.RegisterBankServer(grpcServer, server)
	serverError := grpcServer.Serve(listen)

	if serverError != nil {
		log.Printf("Could not register server")
	}

}

// the bid method in the server checks if the bid placed by the client is a succesful bid, failed bid or an exception
// updates the highestBid and highestBidder value if the placed bid is a success
// returns the Ack enum corresponding to the bid
func (s *Server) Deposit(ctx context.Context, amount *proto.Amount) (*proto.Ack, error) {
	if s.owner == -1 {
		s.owner = amount.Id
	}

	if amount.Amount < 0 {
		log.Println("Can't deposit a negative value!")
		return &proto.Ack{Ack: fail}, nil
	} else if s.owner != amount.Id {
		log.Println("Somebody else is accessing the account!")
		return &proto.Ack{Ack: fail}, nil
	} else {
		s.balance += amount.Amount
		s.owner = amount.Id
		log.Println("Money successfully added to the account")
		return &proto.Ack{Ack: success}, nil
	}

}

func (s *Server) GetBalance(ctx context.Context, in *proto.Empty) (*proto.Balance, error) {
	log.Println("Balance in server was called")
	return &proto.Balance{Balance: s.balance}, nil
}

func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.log", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}

// our enum types
type ack string

const (
	fail    string = "fail"
	success string = "success"
)
