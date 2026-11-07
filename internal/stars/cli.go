package stars

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const helpTemplate = `{{with (or .Long .Short)}}{{.}}

{{end}}{{.UsageString}}`

const usageTemplate = `USAGE
  gh {{.UseLine}}
{{if .HasAvailableSubCommands}}{{$commands := .Commands}}{{range .Groups}}
{{.Title}}
{{$group := .}}{{range $commands}}{{if and .IsAvailableCommand (eq .GroupID $group.ID)}}  {{rpad (printf "%s:" .Name) .NamePadding}} {{.Short}}
{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
{{if .HasParent}}FLAGS{{else}}INHERITED FLAGS{{end}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasAvailableInheritedFlags}}
INHERITED FLAGS
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{with .Annotations.arguments}}
ARGUMENTS
{{.}}
{{end}}{{if .HasExample}}
EXAMPLES
{{.Example}}
{{end}}
LEARN MORE
  Use ` + "`gh stars <command> --help`" + ` for more information about a command.
  Read the documentation at https://github.com/psauxwwf/gh-stars#readme
  Learn about exit codes using ` + "`gh help exit-codes`" + `
`

func Run(ctx context.Context, args []string, out io.Writer) error {
	root := newCommand()
	root.SetArgs(args)
	root.SetOut(out)
	return root.ExecuteContext(ctx)
}

func newCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) error {
		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("initialize GitHub client (authenticate with gh auth login): %w", err)
		}
		switch cmd.Name() {
		case "lists":
			return client.Lists(cmd.Context(), cmd.OutOrStdout())
		case "list":
			return client.List(cmd.Context(), args[0], cmd.OutOrStdout())
		default:
			return client.All(cmd.Context(), cmd.OutOrStdout())
		}
	}
	root := &cobra.Command{
		Use:           "stars <command>",
		Short:         "Work with starred GitHub repositories and star lists.",
		Args:          cobra.NoArgs,
		RunE:          run,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: "  $ gh stars\n" +
			"  $ gh stars lists\n" +
			"  $ gh stars list '/sec/net/tunnel'",
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().BoolP("help", "h", false, "Show help for command")
	root.SetHelpTemplate(helpTemplate)
	root.SetUsageTemplate(usageTemplate)
	root.AddGroup(
		&cobra.Group{ID: "general", Title: "GENERAL COMMANDS"},
		&cobra.Group{ID: "targeted", Title: "TARGETED COMMANDS"},
	)

	root.AddCommand(
		&cobra.Command{
			Use:     "all",
			Short:   "List all starred repository URLs (default)",
			GroupID: "general",
			Args:    cobra.NoArgs,
			RunE:    run,
			Example: "  $ gh stars\n  $ gh stars all",
		},
		&cobra.Command{
			Use:     "lists",
			Short:   "List your star lists",
			GroupID: "general",
			Args:    cobra.NoArgs,
			RunE:    run,
			Example: "  $ gh stars lists",
		},
		&cobra.Command{
			Use:     "list <name>",
			Short:   "List repository URLs from a star list",
			GroupID: "targeted",
			Args: func(cmd *cobra.Command, args []string) error {
				if err := cobra.ExactArgs(1)(cmd, args); err != nil {
					return err
				}
				if args[0] == "" {
					return fmt.Errorf("list name must not be empty")
				}
				return nil
			},
			RunE: run,
			Annotations: map[string]string{
				"arguments": "  Select a list by its exact name, including slashes, spaces, and case.\n" +
					"  Use a list name, not its URL or slug.",
			},
			Example: "  $ gh stars list '/sec/net/tunnel'\n  $ gh stars list 'My favorites'",
		},
	)
	return root
}
