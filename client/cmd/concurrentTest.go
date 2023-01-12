package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"student_22_BFT_Baxos/proto/application"
	"sync"
	"time"
)

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.PersistentFlags().UintVar(&flagLevel, "level", 2, "contention level")
}

// proposeCmd represents the paint command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "concurrent testing",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		wg.Add(2)
		fmt.Println("11111")
		go func() {
			defer wg.Done()

			// connect to instance by checking the address table
			address := cfgInstances["1"]
			fmt.Printf("connecting to %v (%v)\n", flagInstance, address)
			// create connection
			conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		}()

		go func() {
			defer wg.Done()

			// connect to instance by checking the address table
			address := cfgInstances["2"]
			fmt.Printf("connecting to %v (%v)\n", flagInstance, address)
			// create connection
			conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		}()

		fmt.Println("2222")
		wg.Wait()
	},
}
