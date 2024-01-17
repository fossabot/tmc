// Package cli contains implementations of CLI commands. The command code is supposed contain only logic specific to
// the CLI and delegate complex/reusable stuff to code in /internal/commands.
// Commands in cli package should print results in human-readable format to stdout.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/web-of-things-open-source/tm-catalog-cli/internal/model"
)

const filterSep = ","

// Stderrf prints a message to os.Stderr, followed by newline
func Stderrf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	_, _ = fmt.Fprintln(os.Stderr)
}

type FilterFlags struct {
	FilterAuthor       string
	FilterManufacturer string
	FilterMpn          string
	FilterExternalID   string
	Search             string
}

func (ff *FilterFlags) IsSet() bool {
	return ff.FilterAuthor != "" || ff.FilterManufacturer != "" || ff.FilterMpn != "" ||
		ff.FilterExternalID != "" || ff.Search != ""
}

func CreateSearchParamsFromCLI(flags FilterFlags, name string) *model.SearchParams {
	var search *model.SearchParams

	if flags.IsSet() || name != "" {
		search = &model.SearchParams{}
		if flags.FilterAuthor != "" {
			search.Author = strings.Split(flags.FilterAuthor, filterSep)
		}
		if flags.FilterManufacturer != "" {
			search.Manufacturer = strings.Split(flags.FilterManufacturer, filterSep)
		}
		if flags.FilterMpn != "" {
			search.Mpn = strings.Split(flags.FilterMpn, filterSep)
		}
		if flags.FilterExternalID != "" {
			search.ExternalID = strings.Split(flags.FilterExternalID, filterSep)
		}
		if flags.Search != "" {
			search.Query = flags.Search
		}
		if name != "" {
			search.Name = name
		}
	}
	return search
}
