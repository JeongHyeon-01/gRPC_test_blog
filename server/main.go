package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	blogpb "../proto"
)

var db *mongo.Client
var blogdb *mongo.Collection
var mongoCtx context.Context

type BlogServiceServer struct{}

func main() {
	// Configure 'log' package to give file name and line number on eg. log.Fatal
	// Pipe flags to one another (log.LstdFLags = log.Ldate | log.Ltime)

	log.SetFlags(log.LstdFLags | log.Lshortfile)
	fmt.Println("Starting server on port : 50051...")

	//Start our listener, 50051 is default gRPC port
	listener, err := net.Listen("tcp", ":50051")
	//Handle errors if any
	if err != nil {
		log.Fatalf("Unable to listen on port : 50051 : %v", err)
	}

	//set options, here we can configure things like TLS support
	opts := []grpc.ServerOption{}
	//create new gRPC server with (blink) options
	s := grpc.NewServer(opts...)
	// Create Blog Service Type
	srv := &BlogServiceServer{}
	// Register the service with the server
	blogpb.RegisterBlogServiceServer(s, srv)

	// Initialize MongoDb client
	fmt.Println("Connecting to Mongo DB...")

	// non-nil empty context
	mongoCtx = context.Background()

	// Connect takes in a context and options, the connection URI is the only option we pass for now
	db, err = mongo.Connect(mongoCtx, option.Client().ApplyURI("mongodb://localhost:27017"))
	// Handle potential errors
	if err != nil {
		log.Fatal(err)
	}
	// Check whether the connection was succesful by pinging the MongoDB server
	err = db.Ping(mongoCtx, nil)
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v\n", err)
	} else {
		fmt.Println("Connected to Mongodb")
	}
	// Bind our collection to our global variable for use in other methods
	blogdb = db.Database("mydb").Collection("blog")
	// Start the server in a child routine
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	fmt.Println("Server succesfully started on port :50051")

	// Right way to stop the server using a SHUTDOWN HOOK
	// Create a channel to receive OS signals
	c := make(chan os.Signal)

	// Relay os.Interrupt to our channel (os.Interrupt = CTRL+C)
	// Ignore other incoming signals
	signal.Notify(c, os.Interrupt)

	// Block main routine until a signal is received
	// As long as user doesn't press CTRL+C a message is not passed and our main routine keeps running
	<-c

	// After receiving CTRL+C Properly stop the server
	fmt.Println("\nStopping the server...")
	s.Stop()
	listener.Close()
	fmt.Println("Closing MongoDB connection")
	db.Disconnect(mongoCtx)
	fmt.Println("Done.")
}
