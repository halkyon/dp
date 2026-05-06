package cli

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/cache"
	"github.com/halkyon/dp/completion"
	"github.com/halkyon/dp/config"
	"github.com/halkyon/dp/filters"
	"github.com/halkyon/dp/internal/output"
	"github.com/halkyon/dp/internal/shell"
	"github.com/halkyon/dp/server"
	"github.com/halkyon/dp/ssh"
)

var version = ""

func GetVersion() string {
	revision := "unknown"

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				revision = setting.Value
				break
			}
		}
	}

	if version == "" {
		return revision
	}

	if revision == version {
		return version
	}

	return fmt.Sprintf("%s (%s)", version, revision)
}

type CLI struct {
	cfg    *config.Config
	client api.Querier
}

func New(cfg *config.Config, client api.Querier) *CLI {
	return &CLI{cfg: cfg, client: client}
}

type CLICompletionProvider struct {
	client api.Querier
	cfg    *config.Config
}

func NewCLICompletionProvider(client api.Querier, cfg *config.Config) *CLICompletionProvider {
	return &CLICompletionProvider{client: client, cfg: cfg}
}

func (p *CLICompletionProvider) GetCommands() []string {
	var names []string
	for _, cmd := range completionCommands {
		names = append(names, cmd.Name)
	}
	return names
}

func (p *CLICompletionProvider) GetAliases(ctx context.Context) ([]string, error) {
	return p.getFilterList(ctx, filters.NewAliases(p.client, p.cfg.AliasesCache))
}

func (p *CLICompletionProvider) GetLocations(ctx context.Context) ([]string, error) {
	return p.getFilterList(ctx, filters.NewLocations(p.client, p.cfg.LocationsCache))
}

func (p *CLICompletionProvider) GetRegions(ctx context.Context) ([]string, error) {
	return p.getFilterList(ctx, filters.NewRegions(p.client, p.cfg.RegionsCache))
}

func (p *CLICompletionProvider) GetPowerStatuses() []string {
	list, _ := filters.NewPower().Get(context.Background())
	return list
}

func (p *CLICompletionProvider) GetServerStatuses() []string {
	list, _ := filters.NewStatus().Get(context.Background())
	return list
}

func (p *CLICompletionProvider) GetNames(ctx context.Context) ([]string, error) {
	return p.getFilterList(ctx, filters.NewNames(p.client, p.cfg.NamesCache))
}

func (p *CLICompletionProvider) GetTags(ctx context.Context) ([]string, error) {
	return p.getFilterList(ctx, filters.NewTags(p.client, p.cfg.TagsCache))
}

func (p *CLICompletionProvider) GetFields() []string {
	return output.QueryableFields
}

func (p *CLICompletionProvider) getFilterList(ctx context.Context, filter interface {
	Get(context.Context) ([]string, error)
	CacheDuration() time.Duration
	CacheKey() string
}) ([]string, error) {
	var list []string
	var err error

	if filter.CacheDuration() > 0 {
		c, cacheErr := cache.New[[]string](filter.CacheKey(), filter.CacheDuration(), "")
		if cacheErr != nil {
			return nil, cacheErr
		}
		if !c.Get(&list) {
			list, err = filter.Get(ctx)
			if err != nil {
				return nil, err
			}
			if err = c.Set(list, 0); err != nil {
				return nil, err
			}
		}
	} else {
		list, err = filter.Get(ctx)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *CLI) ShowServers(ctx context.Context, opts server.Options, outputFormat string, wide bool) error {
	servers, err := server.List(ctx, c.client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	switch outputFormat {
	case "json":
		encoded, err := output.PrintJSON(servers, opts.Fields)
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
	case "table":
		fmt.Println(output.PrintTable(servers, wide, opts.Fields))
	case "csv":
		w := csv.NewWriter(os.Stdout)
		if err := output.PrintCSV(w, servers, wide, opts.Fields); err != nil {
			return fmt.Errorf("writing CSV: %w", err)
		}
	case "raw":
		fmt.Print(output.PrintRaw(servers, opts.Fields))
	default:
		return fmt.Errorf("unknown output format: %s (use json, table, or csv)", outputFormat)
	}

	return nil
}

func (c *CLI) SSH(ctx context.Context, opts server.Options, sshUser string, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: dp ssh <alias> [ssh flags...]")
	}

	alias := args[0]
	if strings.Contains(alias, "@") {
		alias = strings.SplitN(alias, "@", 2)[1]
	}
	if alias != "" {
		opts.Alias = append(opts.Alias, alias)
	}

	opts.Fields = []string{"Name", "Alias", "IP", "OperatingSystem"}

	servers, err := server.List(ctx, c.client, opts.ToOpts()...)
	if err != nil {
		return err
	}

	return ssh.Run(ctx, servers, sshUser, args)
}

func GenerateCompletion(shellName string) error {
	if shellName == "" {
		detectedShell, err := shell.DetectShell()
		if err != nil {
			return err
		}
		return completion.Generate(detectedShell)
	}
	return completion.Generate(completion.Shell(shellName))
}

func (c *CLI) Filter(ctx context.Context, filterType string) error {
	var filter interface {
		Get(context.Context) ([]string, error)
	}

	switch filterType {
	case "aliases":
		filter = filters.NewAliases(c.client, c.cfg.AliasesCache)
	case "locations":
		filter = filters.NewLocations(c.client, c.cfg.LocationsCache)
	case "regions":
		filter = filters.NewRegions(c.client, c.cfg.RegionsCache)
	case "power":
		filter = filters.NewPower()
	case "status":
		filter = filters.NewStatus()
	case "names":
		filter = filters.NewNames(c.client, c.cfg.NamesCache)
	case "tags":
		filter = filters.NewTags(c.client, c.cfg.TagsCache)
	default:
		return fmt.Errorf("unknown filter type: %s", filterType)
	}

	var list []string
	var err error

	type cacheable interface {
		CacheDuration() time.Duration
		CacheKey() string
	}

	if ca, ok := filter.(cacheable); ok && ca.CacheDuration() > 0 {
		c, err := cache.New[[]string](ca.CacheKey(), ca.CacheDuration(), "")
		if err != nil {
			return err
		}
		if !c.Get(&list) {
			list, err = filter.Get(ctx)
			if err != nil {
				return err
			}
			if err = c.Set(list, 0); err != nil {
				return err
			}
		}
	} else {
		list, err = filter.Get(ctx)
		if err != nil {
			return err
		}
	}

	for _, item := range list {
		fmt.Println(item)
	}
	return nil
}

func Fields() {
	for _, f := range output.QueryableFields {
		fmt.Println(f)
	}
}

var completionCommands = []struct {
	Name        string
	Description string
}{
	{"servers", "List servers with optional filters"},
	{"ssh", "SSH to server by alias"},
	{"completion", "Generate shell completion script"},
	{"generate-completion", "Generate dynamic completion data"},
	{"aliases", "List all server aliases"},
	{"locations", "List all available locations"},
	{"regions", "List all available regions"},
	{"power", "List all power statuses"},
	{"status", "List all server statuses"},
	{"names", "List all server names"},
	{"tags", "List all server tags"},
	{"fields", "List all queryable server fields"},
}

func (c *CLI) GenerateCompletionData(ctx context.Context, filterType, shell, word, prev string) error {
	// If no shell specified, use legacy behavior
	if shell == "" {
		if filterType == "" {
			for _, cmd := range completionCommands {
				fmt.Printf("%s -- %s\n", cmd.Name, cmd.Description)
			}
			return nil
		}

		var filter interface {
			Get(context.Context) ([]string, error)
		}

		switch filterType {
		case "aliases":
			if c.client == nil {
				return nil
			}
			filter = filters.NewAliases(c.client, c.cfg.AliasesCache)
		case "locations":
			if c.client == nil {
				return nil
			}
			filter = filters.NewLocations(c.client, c.cfg.LocationsCache)
		case "regions":
			if c.client == nil {
				return nil
			}
			filter = filters.NewRegions(c.client, c.cfg.RegionsCache)
		case "power":
			filter = filters.NewPower()
		case "status":
			filter = filters.NewStatus()
		case "fields":
			for _, f := range output.QueryableFields {
				fmt.Println(f)
			}
			return nil
		case "names":
			if c.client == nil {
				return nil
			}
			filter = filters.NewNames(c.client, c.cfg.NamesCache)
		case "tags":
			if c.client == nil {
				return nil
			}
			filter = filters.NewTags(c.client, c.cfg.TagsCache)
		default:
			return fmt.Errorf("unknown completion type: %s", filterType)
		}

		list, err := filter.Get(ctx)
		if err != nil {
			return err
		}

		for _, item := range list {
			fmt.Println(item)
		}
		return nil
	}

	// Completion mode: use the filter
	compFilter := completion.NewParser(NewCLICompletionProvider(c.client, c.cfg))

	req := completion.Request{
		Shell: completion.Shell(shell),
		Word:  word,
		Prev:  prev,
	}

	matches, _ := compFilter.Complete(ctx, req)

	for _, match := range matches {
		fmt.Println(match)
	}
	return nil
}
