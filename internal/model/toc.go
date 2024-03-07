package model

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/web-of-things-open-source/tm-catalog-cli/internal/utils"
)

type TOC struct {
	Meta TOCMeta     `json:"meta"`
	Data []*TOCEntry `json:"data"`
}

type TOCMeta struct {
	Created time.Time `json:"created"`
}

type TOCEntry struct {
	Name         string             `json:"name"`
	Manufacturer SchemaManufacturer `json:"schema:manufacturer" validate:"required"`
	Mpn          string             `json:"schema:mpn" validate:"required"`
	Author       SchemaAuthor       `json:"schema:author" validate:"required"`
	Versions     []TOCVersion       `json:"versions"`
}

func (e *TOCEntry) MatchesSearchText(searchQuery string) bool {
	if e == nil {
		return false
	}
	searchQuery = utils.ToTrimmedLower(searchQuery)
	if strings.Contains(utils.ToTrimmedLower(e.Name), searchQuery) {
		return true
	}
	if strings.Contains(utils.ToTrimmedLower(e.Manufacturer.Name), searchQuery) {
		return true
	}
	if strings.Contains(utils.ToTrimmedLower(e.Mpn), searchQuery) {
		return true
	}
	for _, version := range e.Versions {
		if strings.Contains(utils.ToTrimmedLower(version.Description), searchQuery) {
			return true
		}
		if strings.Contains(utils.ToTrimmedLower(version.ExternalID), searchQuery) {
			return true
		}
	}
	return false

}

const TMLinkRel = "content"

type TOCVersion struct {
	Description string            `json:"description"`
	Version     Version           `json:"version"`
	Links       map[string]string `json:"links"`
	TMID        string            `json:"tmID"`
	Digest      string            `json:"digest"`
	TimeStamp   string            `json:"timestamp,omitempty"`
	ExternalID  string            `json:"externalID"`
}

func (toc *TOC) Filter(search *SearchParams) {
	if search == nil {
		return
	}
	toc.Data = slices.DeleteFunc(toc.Data, func(tocEntry *TOCEntry) bool {
		// if !tocEntry.MatchesSearchText(search.Query) {
		// 	return true
		// }

		if !matchesNameFilter(search.Name, tocEntry.Name, search.Options) {
			return true
		}

		if !matchesFilter(search.Author, tocEntry.Author.Name) {
			return true
		}

		if !matchesFilter(search.Manufacturer, tocEntry.Manufacturer.Name) {
			return true
		}

		if !matchesFilter(search.Mpn, tocEntry.Mpn) {
			return true
		}

		return false
	})
	if len(search.Query) > 0 {
		idx, errOpen := bleve.Open("../catalog.bleve")
		if errOpen != nil {
			//return fmt.Errorf("error opening bleve index: %v", errOpen)
		} else {
			defer idx.Close()
			query := bleve.NewQueryStringQuery(search.Query)
			req := bleve.NewSearchRequestOptions(query, 100000, 0, false)
			sr, err := idx.Search(req)
			_ = sr
			if err == nil {
				acceptedValues := make([]string, 0, sr.Size())
				for _, hit := range sr.Hits {
					parts := strings.Split(hit.ID, ":")
					acceptedValues = append(acceptedValues, parts[0])

				}
				fmt.Printf("list from filter %d - list from bleve %d\n", len(toc.Data), len(acceptedValues))
				toc.Data = slices.DeleteFunc(toc.Data, func(tocEntry *TOCEntry) bool {
					return !matchesFilter(acceptedValues, tocEntry.Name)
				})
				fmt.Printf("list after and %d\n", len(toc.Data))
			}
		}
	}
}

func matchesNameFilter(acceptedValue string, value string, options SearchOptions) bool {
	if len(acceptedValue) == 0 {
		return true
	}

	switch options.NameFilterType {
	case FullMatch:
		return value == acceptedValue
	case PrefixMatch:
		actualPathParts := strings.Split(value, "/")
		acceptedValue = strings.Trim(acceptedValue, "/")
		acceptedPathParts := strings.Split(acceptedValue, "/")
		if len(acceptedPathParts) > len(actualPathParts) {
			return false
		}
		return slices.Equal(actualPathParts[0:len(acceptedPathParts)], acceptedPathParts)
	default:
		panic(fmt.Sprintf("unsupported NameFilterType: %d", options.NameFilterType))
	}
}

func matchesFilter(acceptedValues []string, value string) bool {
	if len(acceptedValues) == 0 {
		return true
	}
	return slices.Contains(acceptedValues, value)
}

// findByName searches by name and returns a pointer to the TOCEntry if found
func (toc *TOC) findByName(name string) *TOCEntry {
	for _, value := range toc.Data {
		if value.Name == name {
			return value
		}
	}
	return nil
}

// Insert uses CatalogThingModel to add a version, either to an existing
// entry or as a new entry. Returns the TMID of the inserted entry
func (toc *TOC) Insert(ctm *ThingModel) (TMID, error) {
	tmid, err := ParseTMID(ctm.ID, ctm.IsOfficial())
	if err != nil {
		return TMID{}, err
	}
	// find the right entry, or create if it doesn't exist
	tocEntry := toc.findByName(tmid.Name)
	if tocEntry == nil {
		tocEntry = &TOCEntry{
			Name:         tmid.Name,
			Manufacturer: SchemaManufacturer{Name: tmid.Manufacturer},
			Mpn:          tmid.Mpn,
			Author:       SchemaAuthor{Name: tmid.Author},
		}
		toc.Data = append(toc.Data, tocEntry)
	}
	// TODO: check if id already exists?
	// Append version information to entry
	externalID := ""
	original := ctm.Links.FindLink("original")
	if original != nil {
		externalID = original.HRef
	}
	tv := TOCVersion{
		Description: ctm.Description,
		TimeStamp:   tmid.Version.Timestamp,
		Version:     Version{Model: tmid.Version.Base.String()},
		TMID:        ctm.ID,
		ExternalID:  externalID,
		Digest:      tmid.Version.Hash,
		Links:       map[string]string{"content": tmid.String()},
	}
	tocEntry.Versions = append(tocEntry.Versions, tv)
	return tmid, nil
}

// Delete deletes the record for the given id. Returns TM name to be removed from names file if no more versions are left
func (toc *TOC) Delete(id string) (updated bool, deletedName string, err error) {
	var tocEntry *TOCEntry

	name, found := strings.CutSuffix(id, "/"+filepath.Base(id))
	if !found {
		return false, "", ErrInvalidId
	}
	tocEntry = toc.findByName(name)
	if tocEntry != nil {
		tocEntry.Versions = slices.DeleteFunc(tocEntry.Versions, func(version TOCVersion) bool {
			fnd := version.TMID == id
			if fnd {
				updated = true
			}
			return fnd
		})
		if len(tocEntry.Versions) == 0 {
			toc.Data = slices.DeleteFunc(toc.Data, func(entry *TOCEntry) bool {
				return entry.Name == name
			})
			return updated, name, nil
		}
	}
	return updated, "", nil
}
