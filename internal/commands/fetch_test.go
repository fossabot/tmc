package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/model"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/remotes"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/utils"
)

func TestParseFetchName(t *testing.T) {
	tests := []struct {
		in      string
		expErr  bool
		expName string
		expSV   string
	}{
		{"", true, "", ""},
		{"manufacturer", true, "", ""},
		{"manufacturer\\mpn", true, "", ""},
		{"manu-facturer/mpn", false, "manu-facturer/mpn", ""},
		{"manufacturer/mpn:1.2.3", false, "manufacturer/mpn", "1.2.3"},
		{"manufacturer/mpn:1.2.", true, "", ""},
		{"manufacturer/mpn:1.2", false, "manufacturer/mpn", "1.2"},
		{"manufacturer/mpn:v1.2.3", false, "manufacturer/mpn", "v1.2.3"},
		{"manufacturer/mpn:43748209adcb", true, "", ""},
		{"author/manufacturer/mpn:1.2.3", false, "author/manufacturer/mpn", "1.2.3"},
		{"author/manufacturer/mpn:v1.2.3", false, "author/manufacturer/mpn", "v1.2.3"},
		{"author/manufacturer/mpn/folder/structure:1.2.3", false, "author/manufacturer/mpn/folder/structure", "1.2.3"},
		{"author/manufacturer/mpn/folder/structure:v1.2.3-alpha1", false, "author/manufacturer/mpn/folder/structure", "v1.2.3-alpha1"},
	}

	for _, test := range tests {
		out, err := ParseFetchName(test.in)
		if test.expErr {
			assert.Error(t, err, "Want: error in ParseFetchName(%s). Got: nil", test.in)
			assert.ErrorIs(t, err, ErrInvalidFetchName)
		} else {
			assert.NoError(t, err, "Want: no error in ParseFetchName(%s). Got: %v", test.in, err)
			exp := FetchName{test.expName, test.expSV}
			assert.Equal(t, exp, out, "Want: ParseFetchName(%s) = %v. Got: %v", test.in, exp, out)
		}
	}
}

func TestFetchCommand_FetchByTMIDOrName(t *testing.T) {
	rm := remotes.NewMockRemoteManager(t)
	r := remotes.NewMockRemote(t)
	rm.On("All").Return([]remotes.Remote{r}, nil)
	rm.On("Get", remotes.NewRemoteSpec("r1")).Return(r, nil)
	setUpVersionsForFetchByTMIDOrName(r)

	r.On("Fetch", "manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json").Return("manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", []byte("{\"ver\":\"v1.0.0\"}"), nil)
	r.On("Fetch", "manufacturer/mpn/folder/sub/v1.0.0-20231205123243-c49617d2e4fc.tm.json").Return("manufacturer/mpn/folder/sub/v1.0.0-20231205123243-c49617d2e4fc.tm.json", []byte("{\"ver\":\"v1.0.0\"}"), nil)
	r.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json").Return("author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", []byte("{\"ver\":\"v1.0.0\"}"), nil)
	r.On("Fetch", "author/manufacturer/mpn/v1.0.4-20231206123243-d49617d2e4fc.tm.json").Return("author/manufacturer/mpn/v1.0.4-20231206123243-d49617d2e4fc.tm.json", []byte("{\"ver\":\"v1.0.4\"}"), nil)
	r.On("Fetch", "author/manufacturer/mpn/v1.2.3-20231207153243-e49617d2e4ff.tm.json").Return("author/manufacturer/mpn/v1.2.3-20231207153243-e49617d2e4ff.tm.json", []byte("{\"ver\":\"v1.2.3\"}"), nil)
	r.On("Fetch", "author/manufacturer/mpn/v2.0.0-20231208123243-f49617d2e4fc.tm.json").Return("author/manufacturer/mpn/v2.0.0-20231208123243-f49617d2e4fc.tm.json", []byte("{\"ver\":\"v2.0.0\"}"), nil)
	r.On("Fetch", "author/manufacturer/mpn/folder/sub/v1.0.0-20231205123243-c49617d2e4fc.tm.json").Return("author/manufacturer/mpn/folder/sub/v1.0.0-20231205123243-c49617d2e4fc.tm.json", []byte("{\"ver\":\"v1.0.0\"}"), nil)

	f := NewFetchCommand(rm)

	tests := []struct {
		in         string
		expErr     error
		expErrText string
		expVer     string
	}{
		{"", ErrInvalidFetchName, "must be NAME[:SEMVER]", ""},
		{"manufacturer", ErrInvalidFetchName, "must be NAME[:SEMVER]", ""},
		{"manufacturer/mpn", nil, "", ""},
		{"manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", nil, "", "v1.0.0"},
		{"manufacturer/mpn:v1.0.0", nil, "", "v1.0.0"},
		{"manufacturer/mpn/folder/sub/v1.0.0-20231205123243-c49617d2e4fc.tm.json", nil, "", "v1.0.0"},
		{"author/manufacturer/mpn", nil, "", "v2.0.0"},
		{"author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", nil, "", "v1.0.0"},
		{"author/manufacturer/mpn:v1.0.0", nil, "", "v1.0.0"},
		{"author/manufacturer/mpn:1.0.0", nil, "", "v1.0.0"},
		{"author/manufacturer/mpn:1.a.0", ErrInvalidFetchName, "invalid semantic version", ""},
		{"author/manufacturer/mpn:v1.0", nil, "", "v1.0.4"},
		{"author/manufacturer/mpn:1.3", remotes.ErrTmNotFound, "no version 1.3 found", ""},
		{"author/manufacturer/mpn:1.1", remotes.ErrTmNotFound, "no version 1.1 found", ""},
		{"author/manufacturer/mpn:1.2", nil, "", "v1.2.3"},
		{"author/manufacturer/mpn:3", remotes.ErrTmNotFound, "no version 3 found", ""},
		{"author/manufacturer/mpn:v1", nil, "", "v1.2.3"},
		{"author/manufacturer/mpn/folder/sub", nil, "", "v1.0.0"},
		{"author/manufacturer/mpn/folder/sub:v1.0.0", nil, "", "v1.0.0"},
		{"author/manufacturer/mpn/folder/sub/v1.0.0-20231205123243-c49617d2e4fc.tm.json", nil, "", "v1.0.0"},
	}

	for _, test := range tests {
		_, b, err, _ := f.FetchByTMIDOrName(remotes.EmptySpec, test.in, false)
		if test.expErr != nil {
			assert.ErrorIs(t, err, test.expErr, "Expected error in FetchByTMIDOrName(%s)", test.in)
			assert.ErrorContains(t, err, test.expErrText, "Unexpected error in FetchByTMIDOrName(%s)", test.in)
		} else {
			assert.NoError(t, err, "Expected no error in FetchByTMIDOrName(%s)", test.in)
			assert.True(t, bytes.Contains(b, []byte(test.expVer)), "FetchByTMIDOrName(%s) result does not contain %s. Got: %s", test.in, test.expVer, string(b))
		}
	}
}

func setUpVersionsForFetchByTMIDOrName(r *remotes.MockRemote) {
	r.On("Versions", "manufacturer/mpn").Return([]model.FoundVersion{
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.0.0"},
				TMID:      "manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json",
				Digest:    "c49617d2e4fc",
				TimeStamp: "20231205123243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
	}, nil)
	r.On("Versions", "author/manufacturer/mpn").Return([]model.FoundVersion{
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.0.0"},
				TMID:      "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json",
				Digest:    "c49617d2e4fc",
				TimeStamp: "20231205123243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.0.4"},
				TMID:      "author/manufacturer/mpn/v1.0.4-20231206123243-d49617d2e4fc.tm.json",
				Digest:    "d49617d2e4fc",
				TimeStamp: "20231206123243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.2.0"},
				TMID:      "author/manufacturer/mpn/v1.2.0-20231207163243-e49617d2e4fc.tm.json",
				Digest:    "e49617d2e4fc",
				TimeStamp: "20231207163243", // this is on purpose more recent by timestamp than the latest semver (v.1.2.3)
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.2.1"},
				TMID:      "author/manufacturer/mpn/v1.2.1-20231207133243-e49617d2e4fd.tm.json",
				Digest:    "e49617d2e4fd",
				TimeStamp: "20231207133243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.2.2"},
				TMID:      "author/manufacturer/mpn/v1.2.2-20231207143243-e49617d2e4fe.tm.json",
				Digest:    "e49617d2e4fe",
				TimeStamp: "20231207143243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.2.3"},
				TMID:      "author/manufacturer/mpn/v1.2.3-20231207153243-e49617d2e4ff.tm.json",
				Digest:    "e49617d2e4ff",
				TimeStamp: "20231207153243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v2.0.0"},
				TMID:      "author/manufacturer/mpn/v2.0.0-20231208123243-f49617d2e4fc.tm.json",
				Digest:    "f49617d2e4fc",
				TimeStamp: "20231205123243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
	}, nil)
	r.On("Versions", "author/manufacturer/mpn/folder/sub").Return([]model.FoundVersion{
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.0.0"},
				TMID:      "author/manufacturer/mpn/folder/sub/v1.0.0-20231205123243-c49617d2e4fc.tm.json",
				Digest:    "c49617d2e4fc",
				TimeStamp: "20231205123243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
	}, nil)
}

func TestFetchCommand_FetchByTMIDOrName_MultipleRemotes(t *testing.T) {
	rm := remotes.NewMockRemoteManager(t)
	r1 := remotes.NewMockRemote(t)
	r2 := remotes.NewMockRemote(t)
	rm.On("All").Return([]remotes.Remote{r1, r2}, nil)
	rm.On("Get", remotes.NewRemoteSpec("r1")).Return(r1, nil)
	rm.On("Get", remotes.NewRemoteSpec("r2")).Return(r2, nil)
	r1.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").Return("author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", []byte("{\"src\": \"r1\"}"), nil)
	r1.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json").Return("", []byte{}, remotes.ErrTmNotFound)
	r2.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").Return("", []byte{}, remotes.ErrTmNotFound)
	r2.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json").Return("author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", []byte("{\"src\": \"r2\"}"), nil)
	r1.On("Versions", "author/manufacturer/mpn").Return([]model.FoundVersion{
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.0.0"},
				TMID:      "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json",
				Digest:    "a49617d2e4fc",
				TimeStamp: "20231005123243",
			},
			FoundIn: model.FoundSource{RemoteName: "r1"},
		},
	}, nil)
	r2.On("Versions", "author/manufacturer/mpn").Return([]model.FoundVersion{
		{
			TOCVersion: model.TOCVersion{
				Version:   model.Version{Model: "v1.0.0"},
				TMID:      "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json",
				Digest:    "c49617d2e4fc",
				TimeStamp: "20231205123243",
			},
			FoundIn: model.FoundSource{RemoteName: "r2"},
		},
	}, nil)

	f := NewFetchCommand(rm)
	var id string
	var b []byte
	var err error
	t.Run("fetch from unspecified by id", func(t *testing.T) {
		id, b, err, _ = f.FetchByTMIDOrName(remotes.EmptySpec, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", false)
		assert.NoError(t, err)
		assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", id)
		assert.True(t, bytes.Contains(b, []byte("r1")))

		id, b, err, _ = f.FetchByTMIDOrName(remotes.EmptySpec, "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", false)
		assert.NoError(t, err)
		assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", id)
		assert.True(t, bytes.Contains(b, []byte("r2")))
	})
	t.Run("fetch from named by id", func(t *testing.T) {
		id, b, err, _ = f.FetchByTMIDOrName(remotes.NewRemoteSpec("r1"), "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", false)
		assert.NoError(t, err)
		assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", id)
		assert.True(t, bytes.Contains(b, []byte("r1")))

		id, b, err, _ = f.FetchByTMIDOrName(remotes.NewRemoteSpec("r1"), "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", false)
		assert.Error(t, err)

		id, b, err, _ = f.FetchByTMIDOrName(remotes.NewRemoteSpec("r2"), "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", false)
		assert.Error(t, err)

		id, b, err, _ = f.FetchByTMIDOrName(remotes.NewRemoteSpec("r2"), "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", false)
		assert.NoError(t, err)
		assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", id)
		assert.True(t, bytes.Contains(b, []byte("r2")))
	})
	t.Run("fetch from unspecified by name", func(t *testing.T) {
		id, b, err, _ = f.FetchByTMIDOrName(remotes.EmptySpec, "author/manufacturer/mpn", false)
		assert.NoError(t, err)
		assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", id)
		assert.True(t, bytes.Contains(b, []byte("r2")))

	})
	t.Run("fetch from named by name", func(t *testing.T) {
		id, b, err, _ = f.FetchByTMIDOrName(remotes.NewRemoteSpec("r1"), "author/manufacturer/mpn", false)
		assert.NoError(t, err)
		assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", id)
		assert.True(t, bytes.Contains(b, []byte("r1")))

		id, b, err, _ = f.FetchByTMIDOrName(remotes.NewRemoteSpec("r2"), "author/manufacturer/mpn", false)
		assert.NoError(t, err)
		assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231205123243-c49617d2e4fc.tm.json", id)
		assert.True(t, bytes.Contains(b, []byte("r2")))

	})
}

func TestFetchCommand_FetchByTMID(t *testing.T) {
	rm := remotes.NewMockRemoteManager(t)
	r1 := remotes.NewMockRemote(t)
	r2 := remotes.NewMockRemote(t)
	r1.On("Spec").Return(remotes.NewRemoteSpec("r1"))
	rm.On("All").Return([]remotes.Remote{r1, r2}, nil)

	t.Run("success with unexpected error", func(t *testing.T) {
		r1.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").Return("", nil, errors.New("unexpected")).Once()
		r2.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").Return("author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", []byte("{\"src\": \"r2\"}"), nil).Once()
		f := NewFetchCommand(rm)
		id, b, err, errs := f.FetchByTMID(remotes.EmptySpec, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", false)
		assert.NoError(t, err)
		assert.Len(t, errs, 0)
		assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", id)
		assert.True(t, bytes.Contains(b, []byte("r2")))

	})

	t.Run("not found with unexpected error", func(t *testing.T) {
		r1.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").Return("", nil, errors.New("unexpected")).Once()
		r2.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").Return("", nil, remotes.ErrTmNotFound).Once()
		f := NewFetchCommand(rm)
		_, _, err, errs := f.FetchByTMID(remotes.EmptySpec, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", false)
		assert.ErrorIs(t, err, remotes.ErrTmNotFound)
		if assert.Len(t, errs, 1) {
			assert.ErrorContains(t, errs[0], "unexpected")
		}
	})

}

func TestFetchCommand_FetchByName(t *testing.T) {
	rm := remotes.NewMockRemoteManager(t)
	r1 := remotes.NewMockRemote(t)
	r2 := remotes.NewMockRemote(t)
	rm.On("All").Return([]remotes.Remote{r1, r2}, nil)
	r1Spec := remotes.NewRemoteSpec("r1")
	rm.On("Get", r1Spec).Return(r1, nil)
	r1.On("Spec").Return(r1Spec)
	r2Spec := remotes.NewRemoteSpec("r2")
	//rm.On("Get", r2Spec).Return(r2, nil)
	r2.On("Spec").Return(r2Spec)

	t.Run("name found", func(t *testing.T) {

		r1.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").Return("author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", []byte("{\"src\": \"r1\"}"), nil)
		r1.On("Versions", "author/manufacturer/mpn").Return([]model.FoundVersion{
			{
				TOCVersion: model.TOCVersion{
					Version:   model.Version{Model: "v1.0.0"},
					TMID:      "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json",
					Digest:    "a49617d2e4fc",
					TimeStamp: "20231005123243",
				},
				FoundIn: model.FoundSource{RemoteName: "r1"},
			},
		}, nil)
		r2.On("Versions", "author/manufacturer/mpn").Return(nil, errors.New("unexpected"))

		f := NewFetchCommand(rm)
		t.Run("fetch from unspecified by name", func(t *testing.T) {
			id, b, err, errs := f.FetchByName(remotes.EmptySpec, FetchName{Name: "author/manufacturer/mpn"}, false)
			assert.NoError(t, err)
			if assert.Len(t, errs, 1) {
				assert.ErrorContains(t, errs[0], "unexpected")
			}
			assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", id)
			assert.True(t, bytes.Contains(b, []byte("r1")))

		})
	})
	t.Run("name not found", func(t *testing.T) {

		//r1.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").Return("author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", []byte("{\"src\": \"r1\"}"), nil)
		r1.On("Versions", "author/manufacturer/mpn2").Return(nil, errors.New("unexpected1"))
		r2.On("Versions", "author/manufacturer/mpn2").Return(nil, errors.New("unexpected2"))

		f := NewFetchCommand(rm)
		t.Run("fetch from unspecified by name", func(t *testing.T) {
			_, _, err, errs := f.FetchByName(remotes.EmptySpec, FetchName{Name: "author/manufacturer/mpn2"}, false)
			assert.ErrorIs(t, err, remotes.ErrTmNotFound)
			if assert.Len(t, errs, 2) {
				assert.ErrorContains(t, errs[0], "unexpected1")
				assert.ErrorContains(t, errs[1], "unexpected2")
			}

		})
	})
}

func TestFetchCommand_FetchByTMIDOrName_RestoresId(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expId       string
		expLinksLen int
	}{
		{
			name: "with original id",
			json: `{
  "id": "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json",
  "links": [{
    "rel": "original",
    "href": "externalId"
  }]
}`,
			expId:       "externalId",
			expLinksLen: 0,
		},
		{
			name: "with original id and another link",
			json: `{
  "id": "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json",
  "links": [{
    "rel": "original",
    "href": "externalId"
  },
{
    "rel": "manifest",
    "href": "manifest"
  }]
}`,
			expId:       "externalId",
			expLinksLen: 1,
		},
		{
			name: "without original id",
			json: `{
  "id": "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json",
  "links": [{
    "rel": "manifest",
    "href": "externalId"
  }]
}`,
			expId:       "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json",
			expLinksLen: 1,
		},
		{
			name: "without links at all",
			json: `{
  "id": "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json"
}`,
			expId:       "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json",
			expLinksLen: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rm := remotes.NewMockRemoteManager(t)
			r1 := remotes.NewMockRemote(t)
			r1.On("Versions", "author/manufacturer/mpn").Return([]model.FoundVersion{
				{
					TOCVersion: model.TOCVersion{
						Version:   model.Version{Model: "v1.0.0"},
						TMID:      "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json",
						Digest:    "a49617d2e4fc",
						TimeStamp: "20231005123243",
					},
					FoundIn: model.FoundSource{RemoteName: "r1"},
				},
			}, nil)
			r2 := remotes.NewMockRemote(t)
			r2.On("Versions", "author/manufacturer/mpn").Return(nil, nil)
			rm.On("All").Return([]remotes.Remote{r1, r2}, nil)
			spec := remotes.NewRemoteSpec("r1")
			rm.On("Get", spec).Return(r1, nil)
			f := NewFetchCommand(rm)

			t.Run("with multiple remotes", func(t *testing.T) {
				r1.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").
					Return("author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", []byte(test.json), nil).Once()

				id, b, _, _ := f.FetchByTMIDOrName(remotes.EmptySpec, "author/manufacturer/mpn", true)
				assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", id)
				var js map[string]any
				err := json.Unmarshal(b, &js)
				assert.NoError(t, err)
				jsId := utils.JsGetString(js, "id")
				if assert.NotNil(t, jsId) {
					assert.Equal(t, test.expId, *jsId)
				}
				assert.Len(t, utils.JsGetArray(js, "links"), test.expLinksLen)
			})
			t.Run("with single remote", func(t *testing.T) {
				r1.On("Fetch", "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json").
					Return("author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", []byte(test.json), nil).Once()

				id, b, _, _ := f.FetchByTMIDOrName(spec, "author/manufacturer/mpn", true)
				assert.Equal(t, "author/manufacturer/mpn/v1.0.0-20231005123243-a49617d2e4fc.tm.json", id)
				var js map[string]any
				err := json.Unmarshal(b, &js)
				assert.NoError(t, err)
				jsId := utils.JsGetString(js, "id")
				if assert.NotNil(t, jsId) {
					assert.Equal(t, test.expId, *jsId)
				}
				assert.Len(t, utils.JsGetArray(js, "links"), test.expLinksLen)

			})

		})
	}

}
