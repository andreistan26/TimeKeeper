package main

import (
	"context"
	"fmt"
	"os"

	"github.com/andreistan26/TimeKeeper/cmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "timekeeper",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cmd.SetContext(ctx)
		return nil
	},
}

func main() {
	rootCmd.AddCommand(cmd.CreateServerCommand())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
