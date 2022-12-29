package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"replication/proto"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	id         int
	portNumber int
}

var (
	clientPort   = flag.Int("cPort", 0, "client port number")
	frontendPort = flag.Int("fPort", 0, "frontend port number")
)

func main() {

	flag.Parse()

	client := &Client{
		id:         *clientPort,
		portNumber: *clientPort,
	}

	go connectToFrontend(client)
	log.Printf("Client: %v connected to frontend", client.id)

	for {

	}
}

func connectToFrontend(client *Client) {
	FrontendClient := getFrontendConnection()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		//this is the input of the client who want to make a bid or see the result of the auction
		input := scanner.Text()

		// if the client's input is bid, scans the amount that the client tries to bid and calls the bid() in the frontend
		// if the clietn's input is result, prints the highest bidder and highest bid
		// else prints invalid
		if input == "put" {
			scanner.Scan()
			amountToPut, _ := strconv.ParseInt(scanner.Text(), 10, 0)
			_, err := FrontendClient.Deposit(context.Background(), &proto.Amount{Amount: int32(amountToPut), Id: int32(client.id)})

			if err != nil {
				fmt.Printf("Something is wrong with the bank account")
			}
		} else if input == "balance" {
			balance, _ := FrontendClient.GetBalance(context.Background(), &proto.Empty{})
			fmt.Printf("Balance is: %d \n", balance.Balance)
		} else {
			fmt.Println("Invalid")
		}

	}
}

func getFrontendConnection() proto.BankClient {

	connection, err := grpc.Dial(":"+strconv.Itoa(*frontendPort), grpc.WithTransportCredentials(insecure.NewCredentials())) // remember to put the last line in the dial function

	if err != nil {
		log.Fatalln("Could not dial")
	}

	log.Printf("Dialed")

	return proto.NewBankClient(connection)
}
