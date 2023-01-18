package main

import (
	"flag"
	"fmt"
	"go.dedis.ch/dela/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"os"
	"strconv"
	"student_22_BFT_Baxos/EDDSA"
	"student_22_BFT_Baxos/config"
	"student_22_BFT_Baxos/instance"
	"student_22_BFT_Baxos/polypus/lib/polylib"
	"student_22_BFT_Baxos/proto/BFTBaxos"
	"student_22_BFT_Baxos/proto/application"
)

// bind flag with the corresponding pointer
var configFile = flag.String("config", "../doc/config/node1.yml", "node configuration file")

// Polypus APIs will start at this port. There is one for each node.
const basePolyPort = 4000

func main() {
	//get the input from client cmd line
	flag.Parse()

	cfg, err := config.NewInstanceConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	//polypus
	inWatch := core.NewWatcher()
	outWatch := core.NewWatcher()
	port, _ := strconv.Atoi(cfg.Name)
	polyAddr := fmt.Sprintf("localhost:%d", basePolyPort+port)
	proxy := polylib.Start(polyAddr, inWatch, outWatch, nil)
	polyConfig := fmt.Sprintf(`{"id": "%s", "addr": "%s", "proxy": "http://%s"}`, "A"+cfg.Name, cfg.Listen, proxy)
	fmt.Println(polyConfig)
	// configure information
	//go printConfigure(*cfg)

	//create instance according to configure file
	in := instance.New(cfg.Name, cfg.Increment, cfg.Timeout, cfg.F)

	// add peers
	for i, peer := range cfg.Peers {
		//conn, err := grpc.Dial(peer.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		conn, err := grpc.Dial(peer.Address, polylib.GetDialInterceptor(cfg.Listen, peer.Address, outWatch), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "dial: %v", err)
			os.Exit(1)
		}
		err = in.AddPeer(peer.Name, BFTBaxos.NewBFTBaxosClient(conn))
		if err != nil {
			conn.Close()
			fmt.Fprintf(os.Stderr, "add peer `%v`: %v", peer.Name, err)
			os.Exit(1)
		}

		if i == 3*cfg.F-1 {
			break
		}
	}

	// add private key
	id, err := strconv.Atoi(in.GetName())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to convert ID")
	}
	privatePath := "../" + EDDSA.PrivateKeyPath[id-1]
	pk, err := EDDSA.ReadPrivateKeyFile(privatePath)
	in.KeyStore.PrivateKey = pk
	//fmt.Println("Private Key: ", pk)
	//fmt.Println("Private Key: ", in.KeyStore.PrivateKey)

	// add public key
	for i, path := range EDDSA.PublicKeyPath {
		publicPath := "../" + path
		k, err := EDDSA.ReadPublicKeyFile(publicPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read public key: %v\n", err)
			os.Exit(1)
		}
		//fmt.Println("read public key: ", i)
		//fmt.Println("public key: ", k)
		in.KeyStore.PublicKey[EDDSA.NodeName[i]] = k
	}

	//create server
	//rpcServer := grpc.NewServer()
	serverListenAddr := cfg.Listen
	rpcServer := grpc.NewServer(polylib.GetServerInterceptor(serverListenAddr, inWatch))
	// register service into server
	BFTBaxos.RegisterBFTBaxosServer(rpcServer, in)
	fmt.Println("register consensus service")
	application.RegisterApplicationServer(rpcServer, in)
	fmt.Println("register application service")

	// start listener
	listener, err := net.Listen("tcp", serverListenAddr)
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
