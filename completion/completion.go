package completion

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type Shell string

const (
	ShellBash Shell = "bash"
	ShellZsh  Shell = "zsh"
	ShellFish Shell = "fish"
)

func Generate(shell Shell) error {
	binaryName := getBinaryName()

	switch shell {
	case ShellBash:
		fmt.Print(renderBashCompletion(binaryName))
	case ShellZsh:
		fmt.Print(renderZshCompletion(binaryName))
	case ShellFish:
		fmt.Print(renderFishCompletion(binaryName))
	default:
		return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
	}

	return nil
}

func getBinaryName() string {
	name := os.Args[0]
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

var bashCompletionTmpl = template.Must(template.New("bash").Parse(`# bash completion for {{.Name}}
_dp() {
    local cur prev words cword
    _init_completion || return
    local prev_arg="$prev"
    if [[ $cword -eq 1 ]]; then
        prev_arg=""
    fi
    COMPREPLY=($({{.Name}} generate-completion --shell bash --word "$cur" --prev "$prev_arg" 2>/dev/null))
}

complete -F _dp {{.Name}}
`))

var zshCompletionTmpl = template.Must(template.New("zsh").Parse(`#compdef {{.Name}}
# zsh completion for {{.Name}}
_dp() {
    local cur_word="${words[CURRENT]}"
    local prev_word="${words[CURRENT-1]}"
    if (( CURRENT == 2 )); then
        prev_word=""
    fi
    local -a completions
    while IFS= read -r line; do
        completions+=("${line}")
    done < <({{.Name}} generate-completion --shell zsh --word "$cur_word" --prev "$prev_word" 2>/dev/null)
    _describe "command" completions
}
`))

var fishCompletionTmpl = template.Must(template.New("fish").Parse(`# fish completion for {{.Name}}
function __fish_{{.Name}}_complete
    set -l cmd (commandline -opc)
    set -l cur (commandline -ct)
    set -l prev ""
    if test (count $cmd) -gt 1
        set prev $cmd[-1]
    end
    {{.Name}} generate-completion --shell fish --word "$cur" --prev "$prev" 2>/dev/null
end

complete -c {{.Name}} -f -a '(__fish_{{.Name}}_complete)'
`))

func renderBashCompletion(name string) string {
	var buf strings.Builder
	_ = bashCompletionTmpl.Execute(&buf, struct{ Name string }{Name: name})
	return buf.String()
}

func renderZshCompletion(name string) string {
	var buf strings.Builder
	_ = zshCompletionTmpl.Execute(&buf, struct{ Name string }{Name: name})
	return buf.String()
}

func renderFishCompletion(name string) string {
	var buf strings.Builder
	_ = fishCompletionTmpl.Execute(&buf, struct{ Name string }{Name: name})
	return buf.String()
}
