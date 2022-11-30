package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"student_22_BFT_Baxos/config"
)

var (
	err error

	// the path of configFile
	flagConfigFile string
	// the instance specified by the user
	flagInstance string
	// the color specified by the user
	flagColor string

	// configFile will be read into this file
	cfgQuorum *config.QuorumConfig
	// create cfgInstances name-address mapping
	cfgInstances map[string]string
	// the first instance in the cfgQuorum will be chosen as the cfgDefaultInstance
	cfgDefaultInstance string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "student_22_BFT_Baxos",
	Short: "Client CLI for Colory",
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		fmt.Println("the default quorum file path is: ", flagConfigFile)
		//load the quorum configuration from file
		cfgQuorum, err = config.NewQuorumConfig(flagConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load config: %v\n", err)
			os.Exit(1)
		}

		// create a hashmap for easier access to instance addresses by name
		// also define the default instance name (the first one in the list)
		for i, in := range cfgQuorum.Instances {
			cfgInstances[in.Name] = in.Address
			//define the default instance time
			if i == 0 {
				cfgDefaultInstance = in.Name
			}
		}
	},
}

func init() {
	cfgInstances = make(map[string]string)
	rootCmd.PersistentFlags().StringVar(&flagConfigFile, "config", "doc/config/quorum.yml", "colory quorum configuration file")
	rootCmd.PersistentFlags().StringVar(&flagColor, "color", "green", "specify the color you want to paint")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
