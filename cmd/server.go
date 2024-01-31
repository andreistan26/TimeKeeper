package cmd

import "github.com/spf13/cobra"

func CreateServerCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "server",
		Long: "Starts the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return command
}
