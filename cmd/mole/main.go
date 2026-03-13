package main

import (
	"fmt"
	"os"

	"github.com/Andrei-666/moledrop/internal/words"
	"github.com/spf13/cobra"
)

func main() {
	//starting point of the application
	var rootCmd = &cobra.Command{
		Use:   "moledrop",
		Short: "Moledrop - Ultra-fast P2P file transfer 🦡",
		Long:  `Moledrop is a CLI tool for lightning-fast file sharing between devices in any network`,
	}

	//defining the "send" command
	var sendCmd = &cobra.Command{
		Use:   "send [file]",
		Short: "Generate a unique code for sharing a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file := args[0]
			code := words.GenerateCode()
			fmt.Printf("🦡 Mole is digging the tunnel for: %s\n", file)
			fmt.Printf("Share this code with your friend to receive the file: %s\n", code)
		},
	}

	//defining the "receive" command
	var receiveCmd = &cobra.Command{
		Use:   "receive [code]",
		Short: "Receive a file using the unique code",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			code := args[0]
			fmt.Printf("🦡 Mole is searching for the tunnel with the code: %s\n", code)

		},
	}

	//adding the commands to the root command
	rootCmd.AddCommand(sendCmd, receiveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
