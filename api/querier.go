package api

import "context"

type Querier interface {
	Query(ctx context.Context, query string, variables map[string]any, result any) error
}
