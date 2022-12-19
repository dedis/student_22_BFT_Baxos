package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"os"
	"student_22_BFT_Baxos/colory"
	"student_22_BFT_Baxos/config"
	"student_22_BFT_Baxos/proto/application"
	"student_22_BFT_Baxos/proto/consensus"
)

// bind flag with the corresponding pointer
var configFile = flag.String("config", "doc/config/node1.yml", "Colory configuration file")

func main() {
	//get the input from client cmd line
	flag.Parse()

	cfg, err := config.NewInstanceConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// configure information
	go printConfigure(*cfg)

	//create instance according to configure file
	in := colory.New(cfg.Name, cfg.Increment, cfg.Timeout)

	// add peers
	for _, peer := range cfg.Peers {
		conn, err := grpc.Dial(peer.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "dial: %v", err)
			os.Exit(1)
		}
		err = in.AddPeer(peer.Name, consensus.NewConsensusClient(conn))
		if err != nil {
			conn.Close()
			fmt.Fprintf(os.Stderr, "add peer `%v`: %v", peer.Name, err)
			os.Exit(1)
		}
	}

	//create server
	rpcServer := grpc.NewServer()
	// register service into server
	consensus.RegisterConsensusServer(rpcServer, in)
	fmt.Println("register consensus service")
	application.RegisterApplicationServer(rpcServer, in)
	fmt.Println("register application service")

	// start listener
	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v", err)
		os.Exit(1)
	}
	fmt.Println("start listening")

	err = rpcServer.Serve(listener)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v", err)
		os.Exit(1)
	}
}

func printConfigure(cfg config.InstanceConfig) {
	fmt.Println("-----")
	fmt.Println(cfg.Name)
	fmt.Println(cfg.Increment)
	fmt.Println(cfg.Timeout)
	fmt.Println(cfg.Listen)
	fmt.Println(cfg.Peers)
	fmt.Println("-----")
}
