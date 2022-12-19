/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"student_22_BFT_Baxos/proto/application"
)

// paintCmd represents the paint command
var paintCmd = &cobra.Command{
	Use:   "paint",
	Short: "choose a instance and the color to paint",
	Long:  `chooses the colory instance and input a color you want to paint, it will initiate a consensus process within cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		// select default instance if no one was specified
		if flagInstance == "" {
			flagInstance = cfgDefaultInstance
		}

		// connect to instance by checking the address table
		address := cfgInstances[flagInstance]
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

		// try to paint color
		fmt.Println("painting", flagColor)
		client := application.NewApplicationClient(conn)
		// paint the color from the user
		resp, err := client.Paint(ctx, &application.Request{Color: flagColor})
		if err != nil {
			fmt.Println("The eventual color on the Canvas is ", resp.CanvasColor)
		}
	},
}

func init() {
	rootCmd.AddCommand(paintCmd)
	paintCmd.PersistentFlags().StringVar(&flagInstance, "instance", "", "name of instance to connect to")
}
