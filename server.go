package main

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	pb "srv/proto"

	"google.golang.org/grpc"
)

type messageUnit struct {
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  int
}

type messageQue struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var messageQueObject = messageQue{}

type Server struct{}

func recieveFromStream(csi_ pb.Service_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {
	for {
		req, err := csi_.Recv()
		if err != nil {
			log.Printf("Error reciving request from client :: %v", err)
			errch_ <- err
		} else {
			messageQueObject.mu.Lock()
			messageQueObject.MQue = append(messageQueObject.MQue, messageUnit{ClientName: req.Name, MessageBody: req.Body, MessageUniqueCode: rand.Intn(1e8), ClientUniqueCode: clientUniqueCode_})
			messageQueObject.mu.Unlock()
			log.Printf("%v", messageQueObject.MQue[len(messageQueObject.MQue)-1])
		}

	}
}

func sendToStream(csi_ pb.Service_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {
	for {
		for {
			time.Sleep(500 * time.Millisecond)
			messageQueObject.mu.Lock()
			if len(messageQueObject.MQue) == 0 {
				messageQueObject.mu.Unlock()
				break
			}

			senderUniqueCode := messageQueObject.MQue[0].ClientUniqueCode
			senderName4client := messageQueObject.MQue[0].ClientName
			message4client := messageQueObject.MQue[0].MessageBody
			messageQueObject.mu.Unlock()
			if senderUniqueCode != clientUniqueCode_ {
				err := csi_.Send(&pb.FromServer{Name: senderName4client, Body: message4client})

				if err != nil {
					errch_ <- err
				}

				messageQueObject.mu.Lock()
				if len(messageQueObject.MQue) >= 2 {
					messageQueObject.MQue = messageQueObject.MQue[1:]
				} else {
					messageQueObject.MQue = []messageUnit{}
				}
				messageQueObject.mu.Unlock()
			}
		}
		time.Sleep(100 * time.Microsecond)
	}
}

func (s *Server) ChatService(csi pb.Service_ChatServiceServer) error {
	clientUniqueCode := rand.Intn(1e3)
	errCh := make(chan error)
	// receive request
	go recieveFromStream(csi, clientUniqueCode, errCh)

	// stream >>> client

	go sendToStream(csi, clientUniqueCode, errCh)

	return <-errCh
}

func main() {
	Port := ":5000"
	listen, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", Port, err)
	}
	log.Println("Listening @ " + Port)

	srv := grpc.NewServer()
	cs := Server{}
	pb.RegisterServiceServer(srv, &cs)

	err = srv.Serve(listen)
	if err != nil {
		log.Fatalf("Failed to start gRPC server :: %v", err)
	}

}
