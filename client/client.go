package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	pb "srv/proto"
	"strings"

	"google.golang.org/grpc"
)

type clientHandle struct {
	stream     pb.Service_ChatServiceClient
	clientName string
}

func (ch *clientHandle) clientConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter your name: ")
	msg, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read from console :: %v", err)

	}

	ch.clientName = strings.TrimRight(msg, "\r\n")
}

func (ch *clientHandle) sendMessage() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter your message: ")
		clientMsg, err := reader.ReadString('\n')
		clientMsg = strings.TrimRight(clientMsg, "\r\n")
		if err != nil {
			log.Printf("Failed to read from console :: %v", err)
			continue
		}

		clientMsgBox := &pb.FromClient{
			Name: ch.clientName,
			Body: clientMsg,
		}

		err = ch.stream.Send(clientMsgBox)
		if err != nil {
			log.Printf("Error while sending to server :: %v", err)
		}
	}
}

func (ch *clientHandle) receiveMsg() {
	for {
		resp, err := ch.stream.Recv()
		if err != nil {
			log.Fatalf("can not receive %v", err)
		}
		log.Printf("%s : %s", resp.Name, resp.Body)
	}
}

func main() {
	const serverID = "localhost:5000"
	log.Println("Connecting : " + serverID)

	conn, err := grpc.Dial(serverID, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect gRPC server :: %v", err)
	}

	defer conn.Close()

	client := pb.NewServiceClient(conn)
	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to get response from gRPC server :: %v", err)
	}

	ch := clientHandle{stream: stream}
	ch.clientConfig()

	go ch.sendMessage()
	go ch.receiveMsg()

	bl := make(chan bool)
	<-bl

}
