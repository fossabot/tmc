package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/app/cli"
)

var listCmd = &cobra.Command{
	Use:   "list [PATTERN] [--remote <remoteName>]",
	Short: "List TMs in catalog",
	Long:  `List TMs and filter for PATTERN in all mandatory fields. --remote is optional if there's only one remote configured`,
	Args:  cobra.MaximumNArgs(1),
	Run:   executeList,
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("remote", "r", "", "name of the remote to list")
}

func executeList(cmd *cobra.Command, args []string) {
	remoteName := cmd.Flag("remote").Value.String()

	filter := ""
	if len(args) > 0 {
		filter = args[0]
	}

	err := cli.List(remoteName, filter)
	if err != nil {
		os.Exit(1)
	}
}
