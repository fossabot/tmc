package commands

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/model"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/remotes"
)

func TestVersionsCommand_ListVersions(t *testing.T) {

	t.Run("merged", func(t *testing.T) {

		rm := remotes.NewMockRemoteManager(t)
		r1 := remotes.NewMockRemote(t)
		r2 := remotes.NewMockRemote(t)
		rm.On("All").Return([]remotes.Remote{r1, r2}, nil)
		r1.On("Versions", "senseall").Return(
			[]model.FoundVersion{
				{
					TOCVersion: model.TOCVersion{
						TMID: "omnicorp/senseall/v0.36.0-20231231153548-243d1b462ccc.tm.json",
					},
					FoundIn: model.FoundSource{RemoteName: "r1"},
				},
				{
					TOCVersion: model.TOCVersion{
						TMID: "omnicorp/senseall/v0.35.0-20231230153548-243d1b462bbb.tm.json",
					},
					FoundIn: model.FoundSource{RemoteName: "r1"},
				},
			}, nil)
		r2.On("Versions", "senseall").Return([]model.FoundVersion{
			{
				TOCVersion: model.TOCVersion{
					TMID: "omnicorp/senseall/v0.34.0-20231130153548-243d1b462aaa.tm.json",
				},
				FoundIn: model.FoundSource{RemoteName: "r2"},
			},
			{
				TOCVersion: model.TOCVersion{
					TMID: "omnicorp/senseall/v0.35.0-20231230173548-243d1b462bbb.tm.json",
				},
				FoundIn: model.FoundSource{RemoteName: "r2"},
			},
		}, nil)
		c := NewVersionsCommand(rm)
		res, err, errs := c.ListVersions(remotes.EmptySpec, "senseall")

		assert.NoError(t, err)
		assert.Len(t, errs, 0)
		assert.Len(t, res, 3)
		assert.Equal(t, []model.FoundVersion{
			{
				TOCVersion: model.TOCVersion{TMID: "omnicorp/senseall/v0.34.0-20231130153548-243d1b462aaa.tm.json"},
				FoundIn:    model.FoundSource{RemoteName: "r2"},
			},
			{
				TOCVersion: model.TOCVersion{TMID: "omnicorp/senseall/v0.35.0-20231230173548-243d1b462bbb.tm.json"},
				FoundIn:    model.FoundSource{RemoteName: "r2"},
			},
			{
				TOCVersion: model.TOCVersion{TMID: "omnicorp/senseall/v0.36.0-20231231153548-243d1b462ccc.tm.json"},
				FoundIn:    model.FoundSource{RemoteName: "r1"},
			},
		}, res)

	})

	t.Run("one error", func(t *testing.T) {

		rm := remotes.NewMockRemoteManager(t)
		r1 := remotes.NewMockRemote(t)
		r2 := remotes.NewMockRemote(t)
		r2.On("Spec").Return(remotes.NewRemoteSpec("r2"))
		rm.On("All").Return([]remotes.Remote{r1, r2}, nil)
		r1.On("Versions", "senseall").Return(
			[]model.FoundVersion{
				{
					TOCVersion: model.TOCVersion{
						TMID: "omnicorp/senseall/v0.36.0-20231231153548-243d1b462ccc.tm.json",
					},
					FoundIn: model.FoundSource{RemoteName: "r1"},
				},
				{
					TOCVersion: model.TOCVersion{
						TMID: "omnicorp/senseall/v0.35.0-20231230153548-243d1b462bbb.tm.json",
					},
					FoundIn: model.FoundSource{RemoteName: "r1"},
				},
			}, nil)
		r2.On("Versions", "senseall").Return(nil, errors.New("unexpected error"))
		c := NewVersionsCommand(rm)
		res, err, errs := c.ListVersions(remotes.EmptySpec, "senseall")
		if assert.Len(t, errs, 1) {
			assert.ErrorContains(t, errs[0], "unexpected error")
		}
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, []model.FoundVersion{
			{
				TOCVersion: model.TOCVersion{TMID: "omnicorp/senseall/v0.35.0-20231230153548-243d1b462bbb.tm.json"},
				FoundIn:    model.FoundSource{RemoteName: "r1"},
			},
			{
				TOCVersion: model.TOCVersion{TMID: "omnicorp/senseall/v0.36.0-20231231153548-243d1b462ccc.tm.json"},
				FoundIn:    model.FoundSource{RemoteName: "r1"},
			},
		}, res)

	})

}
