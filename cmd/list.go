package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/web-of-things-open-source/tm-catalog-cli/cmd/completion"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/app/cli"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/remotes"
)

var filterFlags = cli.FilterFlags{}

var listCmd = &cobra.Command{
	Use:   "list <NAME PATTERN>",
	Short: "List TMs in catalog",
	Long: `List TMs in catalog by name pattern, filters or search. 
The name can be a full name or a prefix consisting of complete path parts. 
E.g. 'MyCompany/BarTech' will not match 'MyCompany/BarTechCorp', but will match 'MyCompany/BarTech/BazLamp'.

Name pattern, filters and search can be combined to narrow down the result.`,
	Args:              cobra.MaximumNArgs(1),
	Run:               executeList,
	ValidArgsFunction: completion.CompleteTMNames,
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("remote", "r", "", "name of the remote to list")
	_ = listCmd.RegisterFlagCompletionFunc("remote", completion.CompleteRemoteNames)
	listCmd.Flags().StringP("directory", "d", "", "TM repository directory to list")
	_ = listCmd.MarkFlagDirname("directory")
	listCmd.Flags().StringVar(&filterFlags.FilterAuthor, "filter.author", "", "filter TMs by one or more comma-separated authors")
	listCmd.Flags().StringVar(&filterFlags.FilterManufacturer, "filter.manufacturer", "", "filter TMs by one or more comma-separated manufacturers")
	listCmd.Flags().StringVar(&filterFlags.FilterMpn, "filter.mpn", "", "filter TMs by one or more comma-separated mpn (manufacturer part number)")
	listCmd.Flags().StringVarP(&filterFlags.Search, "search", "s", "", "search TMs by their content matching the search term")
}

func executeList(cmd *cobra.Command, args []string) {
	remoteName := cmd.Flag("remote").Value.String()
	dirName := cmd.Flag("directory").Value.String()

	name := ""
	if len(args) > 0 {
		name = args[0]
	}
	spec, err := remotes.NewSpec(remoteName, dirName)
	if errors.Is(err, remotes.ErrInvalidSpec) {
		cli.Stderrf("Invalid specification of target repository. --remote and --directory are mutually exclusive. Set at most one")
		os.Exit(1)
	}

	search := cli.CreateSearchParamsFromCLI(filterFlags, name)
	err = cli.List(spec, search)
	if err != nil {
		os.Exit(1)
	}
}
