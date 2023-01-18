package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"strconv"
	"student_22_BFT_Baxos/polypus"
	"student_22_BFT_Baxos/polypus/lib/polylib"
	"student_22_BFT_Baxos/proto/application"
	"sync"
	"time"
)

const defaultNumPeers = 3

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.PersistentFlags().UintVar(&flagLevel, "level", 2, "contention level")
}

// proposeCmd represents the paint command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "concurrent testing",
	Run: func(cmd *cobra.Command, args []string) {
		_, outWatch := polypus.InitPoly(3999)

		var wg sync.WaitGroup
		num := defaultNumPeers
		wg.Add(defaultNumPeers)

		for i := 1; i <= num; i++ {
			go func(i int) {
				defer wg.Done()

				// connect to instance by checking the address table
				name := strconv.Itoa(i)
				address := cfgInstances[name]
				fmt.Printf("connecting to %v (%v)\n", flagInstance, address)
				// create connection
				conn, err := grpc.Dial(address, polylib.GetDialInterceptor("client", address, outWatch), grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					fmt.Fprintf(os.Stderr, "dial: %v\n", err)
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
			}(i)
		}
		wg.Wait()
	},
}
