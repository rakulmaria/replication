package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	proto "replication/proto"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Frontend struct {
	proto.UnimplementedBankServer
	name    string
	port    int
	servers []proto.BankClient
	acks    []*proto.Ack
}

var port = flag.Int("port", 0, "server port number") // create the port that recieves the port that the client wants to access to

func main() {
	//setting the log file
	f := setLog()
	defer f.Close()

	flag.Parse()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	frontend := &Frontend{
		name:    "frontend",
		port:    *port,
		servers: make([]proto.BankClient, 0),
		acks:    make([]*proto.Ack, 0),
	}

	go startFrontend(frontend)

	for i := 0; i < 3; i++ {
		conn, err := grpc.Dial("localhost:"+strconv.Itoa(5001+i), grpc.WithTransportCredentials(insecure.NewCredentials()))
		log.Printf("Frontend connected to server at port: %v\n", 5001+i)
		frontend.servers = append(frontend.servers, proto.NewBankClient(conn))
		if err != nil {
			log.Printf("Could not connect: %s", err)
		}
		defer conn.Close()
	}

	for {

	}
}

func (f *Frontend) Deposit(ctx context.Context, bid *proto.Amount) (*proto.Ack, error) {
	// boolean to check if the auction is closed
	f.acks = make([]*proto.Ack, 0)

	log.Printf("Client %d deposits %d to the account", bid.Id, bid.Amount)

	// for each server in the slice, we call the server's bid()
	for index, s := range f.servers {
		ack, err := s.Deposit(ctx, bid)

		// if err != nil the server has crashed and we have to remove it from the slice
		// else the server is still running, add it to the slice
		if err != nil {
			log.Printf("Server crashed")
			f.servers = append(f.servers[:index], f.servers[index+1:]...)
		} else {
			f.acks = append(f.acks, ack)
		}
	}

	ack, err := f.ValidateAcks()

	return ack, err
}

func (f *Frontend) ValidateAcks() (*proto.Ack, error) {
	var sCount = 0
	var fCount = 0

	// counts how many succesful, failed and exception bids there were in the servers
	for i := 0; i < len(f.servers); i++ {
		if f.acks[i].Ack == success {
			sCount++
		}
		if f.acks[i].Ack == fail {
			fCount++
		}
	}

	// checks if more than half of the servers respond were successfull
	// removes the one that was NOT succesful, since it's deprecated
	if sCount > (len(f.servers)/2) && sCount != 0 {
		for i := 0; i < len(f.servers); i++ {
			if f.acks[i].Ack != success {
				// disconnect the server on f.servers[i]
				f.servers = append(f.servers[:i], f.servers[i+1:]...)
			}
		}
		return &proto.Ack{Ack: success}, nil
	}

	// checks if more than half of the servers respond were fail
	// removes the one that was NOT fail, since it's deprecated
	if fCount > (len(f.servers)/2) && fCount != 0 {
		for i := 0; i < len(f.servers); i++ {
			if f.acks[i].Ack != fail {
				// disconnect the server on f.servers[i]
				f.servers = append(f.servers[:i], f.servers[i+1:]...)
			}
		}
		return &proto.Ack{Ack: fail}, nil
	}

	// else everyone answered something different and therefore they're all faulty
	return &proto.Ack{Ack: fail}, errors.New("All the servers are faulty! Run!!")
}

// calls each of the server's Result() and finds the highest bid and bidder
// prints it to the terminal for the client to see
// returns the highestBid and highestBidID
func (f *Frontend) GetBalance(ctx context.Context, in *proto.Empty) (*proto.Balance, error) {
	log.Println("Client asked for the balance")
	balance := int32(0)

	for _, s := range f.servers {
		tmp, _ := s.GetBalance(ctx, in)

		if int32(tmp.Balance) > balance {
			balance = tmp.Balance
		}

	}
	log.Printf("The balance is", balance)
	return &proto.Balance{Balance: balance}, nil
}

func startFrontend(frontend *Frontend) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(frontend.port))

	if err != nil {
		log.Fatalf("Could not create the frontend %v", err)
	}
	log.Printf("Started frontend at port: %d\n", frontend.port)

	// Register the grpc server and serve its listener
	proto.RegisterBankServer(grpcServer, frontend)

	serveError := grpcServer.Serve(listener)
	fmt.Printf("nedern")
	if serveError != nil {
		log.Fatalf("Could not serve listener frontend")

	}
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
