package filters

import (
	"context"
	"sort"
	"time"

	"github.com/halkyon/dp/api"
)

type Tags struct {
	client        api.Querier
	cacheDuration time.Duration
}

func NewTags(client api.Querier, cacheDuration time.Duration) *Tags {
	return &Tags{
		client:        client,
		cacheDuration: cacheDuration,
	}
}

func (t *Tags) CacheDuration() time.Duration { return t.cacheDuration }
func (t *Tags) CacheKey() string             { return "tags" }

func (t *Tags) Get(ctx context.Context) ([]string, error) {
	var tags []string
	tagSet := make(map[string]struct{})

	pageIndex := 0
	pageSize := 50

	input := map[string]any{
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
	}

	for {
		var data serverTagsData
		if err := t.client.Query(ctx, serverTagsQuery, map[string]any{
			"input": input,
		}, &data); err != nil {
			return nil, err
		}

		for _, srv := range data.Servers.Entries {
			for _, tag := range srv.Tags {
				tagSet[tag.Key+"="+tag.Value] = struct{}{}
			}
		}

		if data.Servers.IsLastPage || pageIndex >= data.Servers.PageCount-1 {
			break
		}

		pageIndex++
		input["pageIndex"] = pageIndex
	}

	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	return tags, nil
}

type serverTagNode struct {
	Tags []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"tags"`
}

type serverTagsData struct {
	Servers struct {
		IsLastPage bool            `json:"isLastPage"`
		PageCount  int             `json:"pageCount"`
		Entries    []serverTagNode `json:"entries"`
	} `json:"servers"`
}

const serverTagsQuery = `query($input: PaginatedServersInput) {
	servers(input: $input) {
		isLastPage
		entries {
			tags {
				key
				value
			}
		}
	}
}`
