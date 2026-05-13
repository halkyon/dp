package completion

import (
	"context"
	"strings"
)

type Request struct {
	Shell Shell
	Word  string
	Prev  string
}

type DataProvider interface {
	GetCommands() []string
	GetAliases(ctx context.Context) ([]string, error)
	GetLocations(ctx context.Context) ([]string, error)
	GetRegions(ctx context.Context) ([]string, error)
	GetPowerStatuses() []string
	GetServerStatuses() []string
	GetNames(ctx context.Context) ([]string, error)
	GetTags(ctx context.Context) ([]string, error)
	GetFields() []string
}

type Parser struct {
	dataProvider DataProvider
}

func NewParser(provider DataProvider) *Parser {
	return &Parser{dataProvider: provider}
}

func (e *Parser) Complete(ctx context.Context, req Request) ([]string, error) {
	completionType, word := parseContext(req.Prev, req.Word)
	values := e.getValues(ctx, completionType)
	filtered := filterValues(values, word, completionType)
	return formatForShell(filtered, req.Shell, completionType), nil
}

type completionType string

const (
	compCommands  completionType = "commands"
	compAliases   completionType = "aliases"
	compLocations completionType = "locations"
	compRegions   completionType = "regions"
	compPower     completionType = "power"
	compStatus    completionType = "status"
	compNames     completionType = "names"
	compTags      completionType = "tags"
	compFields    completionType = "fields"
)

func parseContext(prev, word string) (completionType, string) {
	if prev == "" {
		return compCommands, normalizeWord(word)
	}

	// Check for --flag=value patterns
	if idx := strings.Index(prev, "="); idx > 0 {
		flag := prev[:idx]
		value := prev[idx+1:]
		switch flag {
		case "--alias":
			return compAliases, normalizeWord(value + word)
		case "--location":
			return compLocations, normalizeWord(value + word)
		case "--region":
			return compRegions, normalizeWord(value + word)
		case "--power":
			return compPower, normalizeWord(value + word)
		case "--status":
			return compStatus, normalizeWord(value + word)
		case "--name":
			return compNames, normalizeWord(value + word)
		case "--tag":
			return compTags, normalizeWord(value + word)
		case "--query":
			return compFields, normalizeWord(value + word)
		}
	}

	// Check for -fvalue patterns (short flag with value attached)
	if len(prev) >= 2 && prev[0] == '-' && prev[1] != '-' {
		flag := prev[:2]
		value := prev[2:]
		switch flag {
		case "-a":
			return compAliases, normalizeWord(value + word)
		case "-l":
			return compLocations, normalizeWord(value + word)
		case "-r":
			return compRegions, normalizeWord(value + word)
		case "-p":
			return compPower, normalizeWord(value + word)
		case "-s":
			return compStatus, normalizeWord(value + word)
		case "-n":
			return compNames, normalizeWord(value + word)
		case "-t":
			return compTags, normalizeWord(value + word)
		case "-q":
			return compFields, normalizeWord(value + word)
		}
	}

	// Check for exact short/long flags
	switch prev {
	case "--alias", "-a":
		return compAliases, normalizeWord(word)
	case "--location", "-l":
		return compLocations, normalizeWord(word)
	case "--region", "-r":
		return compRegions, normalizeWord(word)
	case "--power", "-p":
		return compPower, normalizeWord(word)
	case "--status", "-s":
		return compStatus, normalizeWord(word)
	case "--name", "-n":
		return compNames, normalizeWord(word)
	case "--tag", "-t":
		return compTags, normalizeWord(word)
	case "--query", "-q":
		return compFields, normalizeWord(word)
	case "ssh":
		return compAliases, normalizeWord(word)
	default:
		return compAliases, normalizeWord(word)
	}
}

func normalizeWord(word string) string {
	// Strip leading quotes
	if len(word) > 0 && (word[0] == '"' || word[0] == '\'') {
		word = word[1:]
	}
	return word
}

func (e *Parser) getValues(ctx context.Context, ct completionType) []string {
	var values []string
	var err error

	switch ct {
	case compCommands:
		values = e.dataProvider.GetCommands()
	case compAliases:
		values, err = e.dataProvider.GetAliases(ctx)
	case compLocations:
		values, err = e.dataProvider.GetLocations(ctx)
	case compRegions:
		values, err = e.dataProvider.GetRegions(ctx)
	case compPower:
		values = e.dataProvider.GetPowerStatuses()
	case compStatus:
		values = e.dataProvider.GetServerStatuses()
	case compNames:
		values, err = e.dataProvider.GetNames(ctx)
	case compTags:
		values, err = e.dataProvider.GetTags(ctx)
	case compFields:
		values = e.dataProvider.GetFields()
	}

	if err != nil {
		return []string{}
	}
	return values
}

func filterValues(values []string, word string, ct completionType) []string {
	if word == "" {
		return values
	}

	// Handle @ prefix for aliases
	if ct == compAliases {
		if idx := strings.LastIndex(word, "@"); idx >= 0 {
			word = word[idx+1:]
		}
	}

	wordLower := strings.ToLower(word)
	var result []string
	for _, v := range values {
		if strings.HasPrefix(strings.ToLower(v), wordLower) {
			result = append(result, v)
		}
	}
	return result
}

func formatForShell(values []string, shell Shell, ct completionType) []string {
	switch shell {
	case ShellZsh:
		return formatZsh(values, ct)
	case ShellFish:
		return formatFish(values, ct)
	default:
		// bash or unknown
		return values
	}
}

func formatZsh(values []string, ct completionType) []string {
	result := make([]string, len(values))
	switch ct {
	case compRegions:
		for i, v := range values {
			result[i] = v + ":" + regionTitle(v)
		}
	case compLocations, compFields:
		for i, v := range values {
			result[i] = v
		}
	default:
		for i, v := range values {
			result[i] = v + ":" + v
		}
	}
	return result
}

func regionTitle(region string) string {
	titles := map[string]string{
		"AP": "Asia Pacific",
		"EU": "Europe",
		"NA": "North America",
		"SA": "South America",
	}
	if title, ok := titles[region]; ok {
		return title
	}
	return region
}

func formatFish(values []string, ct completionType) []string {
	result := make([]string, len(values))
	switch ct {
	case compRegions:
		for i, v := range values {
			result[i] = v + "\t" + regionTitle(v)
		}
	case compLocations, compFields:
		for i, v := range values {
			result[i] = v
		}
	default:
		for i, v := range values {
			result[i] = v + "\t" + v
		}
	}
	return result
}
