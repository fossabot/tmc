package cli

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/web-of-things-open-source/tm-catalog-cli/internal/commands"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/model"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/remotes"
)

// TODO: figure out how to use viper
const columnWidthName = "TMC_COLUMNWIDTH"
const columnWidthDefault = 40

func List(remoteName, filter string) error {
	toc, err := commands.NewListCommand(remotes.DefaultManager()).List(remoteName, &model.SearchParams{
		Query: filter,
	})
	if err != nil {
		Stderrf("Error listing: %v", err)
		return err
	}
	printToC(toc, filter)
	return nil
}

// TODO: use better table writer with eliding etc.
func printToC(toc model.SearchResult, filter string) {
	colWidth := columnWidth()
	table := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintf(table, "MANUFACTURER\tMODEL\tNAME\n")
	for _, value := range toc.Entries {
		name := value.Name
		man := elideString(value.Manufacturer.Name, colWidth)
		mdl := elideString(value.Mpn, colWidth)
		_, _ = fmt.Fprintf(table, "%s\t%s\t%s\n", man, mdl, name)
	}
	_ = table.Flush()
}

func elideString(value string, colWidth int) string {
	if len(value) < colWidth {
		return value
	}

	var elidedValue string
	for i, rn := range value {
		elidedValue += string(rn)
		if i >= (colWidth - 4) {
			return elidedValue + "..."
		}
	}
	return value + "..."
}

func columnWidth() int {
	cw, err := strconv.Atoi(os.Getenv(columnWidthName))
	if err != nil {
		cw = columnWidthDefault
	}
	return cw
}
