package cmd

import (
	"errors"
	"os"

	"github.com/spf13/viper"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/config"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/remotes"

	"github.com/spf13/cobra"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/app/cli"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve TMs in catalog",
	Long: `Serve TMs in catalog.
A target for push operations must be specified with either --directory or --remote.
This may be omitted if there's exactly one remote configured`,
	Args: cobra.MaximumNArgs(0),
	Run:  serve,
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("host", "", "0.0.0.0", "serve with this host name")
	serveCmd.Flags().StringP("port", "", "8080", "serve with this port")
	serveCmd.Flags().StringP("remote", "r", "", "name of the remote target for push")
	serveCmd.Flags().StringP("directory", "d", "", "TM repository directory for push")
	serveCmd.Flags().StringP("urlContextRoot", "", "",
		"define additional URL context root path to be considered in hypermedia links,\ncan also be set via environment variable TMC_URLCONTEXTROOT")
	_ = viper.BindPFlag(config.KeyUrlContextRoot, serveCmd.Flags().Lookup("urlContextRoot"))
}

func serve(cmd *cobra.Command, args []string) {
	host := cmd.Flag("host").Value.String()
	port := cmd.Flag("port").Value.String()
	remote := cmd.Flag("remote").Value.String()
	dir := cmd.Flag("directory").Value.String()
	spec, err := remotes.NewSpec(remote, dir)
	if errors.Is(err, remotes.ErrInvalidSpec) {
		cli.Stderrf("Invalid specification of target repository. --remote and --directory are mutually exclusive. Set at most one")
		os.Exit(1)
	}

	urlCtxRoot := viper.GetString(config.KeyUrlContextRoot)

	err = cli.Serve(host, port, urlCtxRoot, spec)
	if err != nil {
		cli.Stderrf("serve failed")
		os.Exit(1)
	}
}
