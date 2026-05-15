package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/config"
	"github.com/halkyon/dp/internal/cli"
	"github.com/halkyon/dp/server"
)

var (
	outputFormatFlag string
	outputWide       bool
	queryFields      = new(stringSlice)

	flagName     = new(stringSlice)
	flagAlias    = new(stringSlice)
	flagLocation = new(stringSlice)
	flagRegion   = new(stringSlice)
	flagStatus   = new(stringSlice)
	flagPower    = new(stringSlice)
	flagTag      = new(stringSlice)
	flagUser     = new(stringSlice)

	sortField string

	completionShell string
	completionWord  string
	completionPrev  string
)

type stringSlice []string

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func init() {
	flag.BoolFunc("V", "Print version and exit", func(string) error {
		fmt.Println(cli.GetVersion())
		os.Exit(0)
		return nil
	})
	flag.BoolFunc("version", "Print version and exit", func(string) error {
		fmt.Println(cli.GetVersion())
		os.Exit(0)
		return nil
	})

	flag.StringVar(&outputFormatFlag, "o", "", "Output format (shorthand): json, table, csv")
	flag.StringVar(&outputFormatFlag, "output", "", "Output format: json, table, csv")
	flag.BoolVar(&outputWide, "ow", false, "Show more fields (shorthand)")
	flag.Bool("output-wide", false, "Show more fields in table/csv output")
	flag.Var(queryFields, "q", "Output specific field(s) (repeatable)")
	flag.Var(queryFields, "query", "Output specific field(s) (repeatable)")

	flag.Var(flagName, "n", "Filter by name (repeatable)")
	flag.Var(flagName, "name", "Filter by name (repeatable)")
	flag.Var(flagAlias, "a", "Filter by alias (repeatable)")
	flag.Var(flagAlias, "alias", "Filter by alias (repeatable)")
	flag.Var(flagLocation, "l", "Filter by location (repeatable)")
	flag.Var(flagLocation, "location", "Filter by location (repeatable)")
	flag.Var(flagRegion, "r", "Filter by region (repeatable)")
	flag.Var(flagRegion, "region", "Filter by region (repeatable)")
	flag.Var(flagStatus, "s", "Filter by status (repeatable)")
	flag.Var(flagStatus, "status", "Filter by status (repeatable)")
	flag.Var(flagPower, "p", "Filter by power status (repeatable)")
	flag.Var(flagPower, "power", "Filter by power status (repeatable)")
	flag.Var(flagTag, "t", "Filter by tag (repeatable)")
	flag.Var(flagTag, "tag", "Filter by tag (repeatable)")

	flag.StringVar(&sortField, "S", "", "Sort by field")
	flag.StringVar(&sortField, "sort", "", "Sort by field")

	flag.Var(flagUser, "user", "SSH user (for ssh command)")

	// generate-completion flags
	flag.StringVar(&completionShell, "shell", "", "Shell type for completion (bash, zsh, fish)")
	flag.StringVar(&completionWord, "word", "", "Current word being completed")
	flag.StringVar(&completionPrev, "prev", "", "Previous word on command line")
}

func main() {
	flag.Usage = func() {
		cmd := ""
		args := os.Args[1:]
		for _, a := range args {
			if !strings.HasPrefix(a, "-") {
				cmd = a
				break
			}
		}

		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command> [command options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  servers             List servers with optional filters\n")
		fmt.Fprintf(os.Stderr, "  ssh                 SSH to server by alias\n")
		fmt.Fprintf(os.Stderr, "  completion          Generate shell completion script\n")
		fmt.Fprintf(os.Stderr, "  generate-completion Generate dynamic completion data\n")
		fmt.Fprintf(os.Stderr, "  aliases             List all server aliases\n")
		fmt.Fprintf(os.Stderr, "  locations           List all available locations\n")
		fmt.Fprintf(os.Stderr, "  regions             List all available regions\n")
		fmt.Fprintf(os.Stderr, "  power               List all power statuses\n")
		fmt.Fprintf(os.Stderr, "  status              List all server statuses\n")
		fmt.Fprintf(os.Stderr, "  names               List all server names\n")
		fmt.Fprintf(os.Stderr, "  tags                List all server tags\n")
		fmt.Fprintf(os.Stderr, "  fields              List all queryable server fields\n\n")

		switch cmd {
		case "servers":
			fmt.Fprintf(os.Stderr, "Options for servers:\n")
			fmt.Fprintf(os.Stderr, "  -o, --output <format>   Output format: json, table, csv (default json)\n")
			fmt.Fprintf(os.Stderr, "  -ow, --output-wide      Show more fields in table/csv\n")
			fmt.Fprintf(os.Stderr, "  -q, --query <field>     Output specific field(s) (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -S, --sort <field>      Sort by field\n")
			fmt.Fprintf(os.Stderr, "\nFilters (shorthand, long form):\n")
			fmt.Fprintf(os.Stderr, "  -n, --name <val>      Filter by name (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -a, --alias <val>     Filter by alias (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -l, --location <val>  Filter by location (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -r, --region <val>    Filter by region (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -s, --status <val>    Filter by status (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -p, --power <val>     Filter by power status (repeatable)\n")
			fmt.Fprintf(os.Stderr, "  -t, --tag <val>       Filter by tag (repeatable)\n")
		case "ssh":
			fmt.Fprintf(os.Stderr, "Options for ssh:\n")
			fmt.Fprintf(os.Stderr, "  --user <name>           SSH user (default: root on linux, admin on windows)\n")
		case "completion":
			fmt.Fprintf(os.Stderr, "Options for completion:\n")
			fmt.Fprintf(os.Stderr, "  [shell name] (optional: bash, zsh, fish; auto-detected if omitted)\n")
		default:
		}
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	cmdArgs := flag.Args()[1:]

	if len(cmdArgs) > 0 {
		if err := flag.CommandLine.Parse(cmdArgs); err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing flags:", err)
			os.Exit(1)
		}
	}

	opts := server.Options{
		Name:     *flagName,
		Alias:    *flagAlias,
		Location: *flagLocation,
		Region:   *flagRegion,
		Status:   *flagStatus,
		Power:    *flagPower,
		Tag:      *flagTag,
		Sort:     sortField,
	}

	// After re-parsing, remaining non-flag args are in flag.Args()
	var remainingArgs []string
	if len(cmdArgs) > 0 {
		remainingArgs = flag.Args()
	}

	if err := run(cmd, remainingArgs, opts); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func validateCmdFlags(cmd string) error {
	if cmd != "servers" {
		hasFilter := len(*flagName) > 0 || len(*flagAlias) > 0 || len(*flagLocation) > 0 ||
			len(*flagRegion) > 0 || len(*flagStatus) > 0 || len(*flagPower) > 0 ||
			len(*flagTag) > 0 || len(*queryFields) > 0 || outputFormatFlag != "" ||
			outputWide || flag.Lookup("output-wide").Value.String() == "true" ||
			sortField != ""
		if hasFilter {
			return errors.New("output and filter flags are only valid for the servers command")
		}
	}
	if cmd != "ssh" && len(*flagUser) > 0 {
		return errors.New("flag --user is only valid for the ssh command")
	}
	return nil
}

func run(cmd string, args []string, opts server.Options) error {
	if err := validateCmdFlags(cmd); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var client api.Querier
	if cfg.APIKey != "" {
		c, err := api.NewClient(cfg.APIKey)
		if err != nil {
			return err
		}
		if cfg.APIURL != "" {
			c.SetBaseURL(cfg.APIURL)
		}
		client = c
	}

	needsClient := cmd == "servers" || cmd == "ssh" || cmd == "aliases" || cmd == "locations" || cmd == "regions" || cmd == "names" || cmd == "tags"
	if needsClient && client == nil {
		return errors.New("API key required; set DATAPACKET_API_KEY or api_key in ~/.config/dp/credentials")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	c := cli.New(cfg, client)

	switch cmd {
	case "version":
		fmt.Println(cli.GetVersion())
		return nil
	case "fields":
		cli.Fields()
		return nil
	case "completion":
		shellName := ""
		if len(args) > 0 {
			shellName = args[0]
		}
		return cli.GenerateCompletion(shellName)
	case "servers":
		wide := outputWide || flag.Lookup("output-wide").Value.String() == "true"
		output := "json"
		if cfg.Output != "" {
			output = cfg.Output
		}
		if outputFormatFlag != "" {
			output = outputFormatFlag
		}
		if len(*queryFields) > 0 {
			opts.Fields = *queryFields
		}
		return c.ShowServers(ctx, opts, output, wide)
	case "ssh":
		sshUser := ""
		if len(*flagUser) > 0 {
			sshUser = (*flagUser)[0]
		}
		if len(args) < 1 {
			return errors.New("usage: dp ssh <alias> [ssh flags...]")
		}
		return c.SSH(ctx, opts, sshUser, args)
	case "aliases":
		return c.Filter(ctx, "aliases")
	case "locations":
		return c.Filter(ctx, "locations")
	case "regions":
		return c.Filter(ctx, "regions")
	case "power":
		return c.Filter(ctx, "power")
	case "status":
		return c.Filter(ctx, "status")
	case "names":
		return c.Filter(ctx, "names")
	case "tags":
		return c.Filter(ctx, "tags")
	case "generate-completion":
		filterType := ""
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			filterType = args[0]
		}
		return c.GenerateCompletionData(ctx, filterType, completionShell, completionWord, completionPrev)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}
