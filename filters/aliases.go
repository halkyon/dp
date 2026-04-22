package filters

import (
	"context"

	"github.com/halkyon/dp/api"
)

type Aliases struct {
	client api.Querier
}

func NewAliases(client api.Querier) *Aliases {
	return &Aliases{
		client: client,
	}
}

type serverAliasNode struct {
	Alias string `json:"alias"`
}

type serverAliasesData struct {
	Servers struct {
		IsLastPage bool              `json:"isLastPage"`
		Entries    []serverAliasNode `json:"entries"`
	} `json:"servers"`
}

const serverAliasesQuery = `query($input: PaginatedServersInput) {
	servers(input: $input) {
		isLastPage
		entries {
			alias
		}
	}
}`

func (a *Aliases) Get(ctx context.Context) ([]string, error) {
	var aliases []string

	pageIndex := 0
	pageSize := 50

	input := map[string]any{
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
	}

	for {
		var data serverAliasesData
		if err := a.client.Query(ctx, serverAliasesQuery, map[string]any{
			"input": input,
		}, &data); err != nil {
			return nil, err
		}

		for _, srv := range data.Servers.Entries {
			if srv.Alias != "" {
				aliases = append(aliases, srv.Alias)
			}
		}

		if data.Servers.IsLastPage {
			break
		}

		pageIndex++
		input["pageIndex"] = pageIndex
	}

	return aliases, nil
}
