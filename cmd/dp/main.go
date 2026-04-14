package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/completion"
	"github.com/halkyon/dp/config"
	"github.com/halkyon/dp/server"
	"github.com/halkyon/dp/ssh"
)

var version = ""
var verbose = false
var sshUser = ""
var serverOpts server.Options

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	parseFlags()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func parseFlags() {
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--verbose", "-v":
			verbose = true
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			i--
		case "--user", "-u":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: --user requires a value")
				os.Exit(1)
			}
			sshUser = os.Args[i+1]
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
		case "--name", "-n":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: --name requires a value")
				os.Exit(1)
			}
			serverOpts.Name = append(serverOpts.Name, os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
		case "--alias", "-a":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: --alias requires a value")
				os.Exit(1)
			}
			serverOpts.Alias = append(serverOpts.Alias, os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
		case "--location", "-l":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: --location requires a value")
				os.Exit(1)
			}
			serverOpts.Location = append(serverOpts.Location, os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
		case "--region", "-r":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: --region requires a value")
				os.Exit(1)
			}
			serverOpts.Region = append(serverOpts.Region, os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
		case "--status":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: --status requires a value")
				os.Exit(1)
			}
			serverOpts.Status = append(serverOpts.Status, os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
		case "--power":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: --power requires a value")
				os.Exit(1)
			}
			serverOpts.Power = append(serverOpts.Power, os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
		case "--tag", "-t":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: --tag requires a value")
				os.Exit(1)
			}
			serverOpts.Tag = append(serverOpts.Tag, os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		}
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	switch os.Args[1] {
	case "show":
		return runShow(ctx)
	case "ssh":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: dp ssh <alias> [ssh flags...]")
		}
		return runSSH(ctx, os.Args[2:])
	case "completion":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: dp completion <bash|zsh|fish>")
		}
		return completion.Generate(completion.Shell(os.Args[2]))
	case "aliases":
		return completion.ListAliases(ctx)
	case "version", "-V", "--version":
		fmt.Println(getVersion())
		return nil
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] <command> [options]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  -v, --verbose      Print verbose information\n")
	fmt.Fprintf(os.Stderr, "  -u, --user <user>  SSH user (for ssh command)\n")
	fmt.Fprintf(os.Stderr, "Global filter options (for show, ssh, aliases):\n")
	fmt.Fprintf(os.Stderr, "  -n, --name <name>        Filter by name (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  -a, --alias <alias>     Filter by alias (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  -l, --location <loc>    Filter by location (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  -r, --region <region>   Filter by region (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  --status <status>       Filter by status (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  --power <power>         Filter by power status (repeatable)\n")
	fmt.Fprintf(os.Stderr, "  -t, --tag <key=value>    Filter by tag (repeatable)\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  show [regex]       List servers (optional regex filter)\n")
	fmt.Fprintf(os.Stderr, "  ssh <alias>        SSH to server (alias or user@alias) [ssh flags...]\n")
	fmt.Fprintf(os.Stderr, "  completion <shell> Generate completion script (bash|zsh|fish)\n")
	fmt.Fprintf(os.Stderr, "  aliases            List all server aliases (for completion)\n")
	fmt.Fprintf(os.Stderr, "  version            Print version\n")
}

func getVersion() string {
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

func runShow(ctx context.Context) error {
	opts := serverOpts

	for i := 2; i < len(os.Args); i++ {
		if !strings.HasPrefix(os.Args[i], "-") {
			opts.Filter = os.Args[i]
		}
	}

	apiKey, err := config.GetAPIKey()
	if err != nil {
		return err
	}
	if apiKey == "" {
		return api.ErrMissingAPIKey
	}

	client, err := api.NewClient(apiKey)
	if err != nil {
		return err
	}

	servers, err := server.FetchAll(ctx, client, opts)
	if err != nil {
		return err
	}

	servers = server.Filter(servers, opts)

	output, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	fmt.Println(string(output))

	return nil
}

func runSSH(ctx context.Context, args []string) error {
	servers, err := server.FetchAll(ctx, getClient(), serverOpts)
	if err != nil {
		return err
	}

	return ssh.Run(ctx, servers, sshUser, args, verbose)
}

func getClient() *api.Client {
	apiKey, _ := config.GetAPIKey()
	client, _ := api.NewClient(apiKey)
	return client
}
