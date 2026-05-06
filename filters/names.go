package filters

import (
	"context"
	"sort"
	"time"

	"github.com/halkyon/dp/api"
)

type Names struct {
	client        api.Querier
	cacheDuration time.Duration
}

func NewNames(client api.Querier, cacheDuration time.Duration) *Names {
	return &Names{
		client:        client,
		cacheDuration: cacheDuration,
	}
}

func (n *Names) CacheDuration() time.Duration { return n.cacheDuration }
func (n *Names) CacheKey() string             { return "names" }

func (n *Names) Get(ctx context.Context) ([]string, error) {
	var names []string
	nameSet := make(map[string]struct{})

	pageIndex := 0
	pageSize := 50

	input := map[string]any{
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
	}

	for {
		var data serverNamesData
		if err := n.client.Query(ctx, serverNamesQuery, map[string]any{
			"input": input,
		}, &data); err != nil {
			return nil, err
		}

		for _, srv := range data.Servers.Entries {
			if srv.Name != "" {
				nameSet[srv.Name] = struct{}{}
			}
		}

		if data.Servers.IsLastPage || pageIndex >= data.Servers.PageCount-1 {
			break
		}

		pageIndex++
		input["pageIndex"] = pageIndex
	}

	for name := range nameSet {
		names = append(names, name)
	}
	sort.Strings(names)

	return names, nil
}

type serverNameNode struct {
	Name string `json:"name"`
}

type serverNamesData struct {
	Servers struct {
		IsLastPage bool             `json:"isLastPage"`
		PageCount  int              `json:"pageCount"`
		Entries    []serverNameNode `json:"entries"`
	} `json:"servers"`
}

const serverNamesQuery = `query($input: PaginatedServersInput) {
	servers(input: $input) {
		isLastPage
		entries {
			name
		}
	}
}`
