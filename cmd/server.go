package cmd

import (
	"github.com/andreistan26/TimeKeeper/server"
	"github.com/spf13/cobra"
)

func CreateServerCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "server",
		Long: "Starts the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := server.StartServer(cmd.Context()); err != nil {
				return err
			}
			return nil
		},
	}

	return command
}
