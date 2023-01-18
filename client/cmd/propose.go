package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"student_22_BFT_Baxos/polypus"
	"student_22_BFT_Baxos/polypus/lib/polylib"
	"student_22_BFT_Baxos/proto/application"
	"time"
)

func init() {
	rootCmd.AddCommand(proposeCmd)
	proposeCmd.PersistentFlags().StringVar(&flagInstance, "instance", "", "name of instance to connect to")
}

// proposeCmd represents the paint command
var proposeCmd = &cobra.Command{
	Use:   "propose",
	Short: "choose a instance and the value to propose",
	Long:  `chooses the instance instance and input a value you want to propose, it will initiate a consensus process within cluster`,
	Run: func(cmd *cobra.Command, args []string) {

		_, outWatch := polypus.InitPoly(3999)
		// select default instance if no one was specified
		if flagInstance == "" {
			flagInstance = cfgDefaultInstance
		}
		fmt.Println("cfgDefaultInstance: " + cfgDefaultInstance)
		fmt.Println("flag: " + flagInstance)
		// connect to instance by checking the address table
		address := cfgInstances[flagInstance]
		fmt.Printf("connecting to %v (%v)\n", flagInstance, address)
		// create connection
		conn, err := grpc.Dial(address, polylib.GetDialInterceptor("client", address, outWatch), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "dial: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), cfgQuorum.Timeout)
		defer cancel()

		// try to propose value
		fmt.Println("propsoe value: ", flagValue)
		client := application.NewApplicationClient(conn)
		startTime := time.Now()
		// propose the value from the user
		resp, err := client.Propose(ctx, &application.Request{Color: flagValue})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(resp.Result)
		fmt.Println("take ", time.Now().Sub(startTime), " to reach consensus")
	},
}
