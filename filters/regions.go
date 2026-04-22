package filters

import (
	"context"
	"sort"

	"github.com/halkyon/dp/api"
)

type Regions struct {
	client api.Querier
}

func NewRegions(client api.Querier) *Regions {
	return &Regions{
		client: client,
	}
}

func (r *Regions) Get(ctx context.Context) ([]string, error) {
	var data locationsData
	if err := r.client.Query(ctx, locationsQuery, nil, &data); err != nil {
		return nil, err
	}

	regionSet := make(map[string]struct{})
	for _, loc := range data.Locations {
		regionSet[loc.Region] = struct{}{}
	}

	regions := make([]string, 0, len(regionSet))
	for region := range regionSet {
		regions = append(regions, region)
	}
	sort.Strings(regions)

	return regions, nil
}
