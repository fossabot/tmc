package commands

import (
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/model"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/remotes"
)

type ListCommand struct {
}

func NewListCommand() *ListCommand {
	return &ListCommand{}
}
func (c *ListCommand) List(rSpec model.RepoSpec, search *model.SearchParams) (model.SearchResult, error, []*remotes.RepoAccessError) {
	rs, err := remotes.GetSpecdOrAll(rSpec)
	if err != nil {
		return model.SearchResult{}, err, nil
	}
	sr, errs := rs.List(search)
	return sr, nil, errs
}
