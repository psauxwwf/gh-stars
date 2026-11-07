package stars

import (
	"context"
	"fmt"
	"io"
)

const usage = `Usage: gh stars [all | lists | list <name>]

  all          Print all starred repository URLs (default)
  lists        Print the names of your star lists
  list <name>  Print repository URLs from an exactly matching list
  -h, --help   Show this help

Examples:
  gh stars
  gh stars lists
  gh stars list '/sec/net/tunnel'

Authentication: gh auth login, GH_TOKEN, or GITHUB_TOKEN.
Output is always plain text, one result per line, without a pager.
`

func Run(ctx context.Context, args []string, out io.Writer) error {
	command := "all"
	if len(args) > 0 {
		command = args[0]
	}
	switch command {
	case "-h", "--help", "help":
		_, err := io.WriteString(out, usage)
		return err
	case "all", "lists":
		if len(args) > 1 {
			return fmt.Errorf("%s takes no arguments; see gh stars --help", command)
		}
	case "list":
		if len(args) != 2 || args[1] == "" {
			return fmt.Errorf("usage: gh stars list <name>")
		}
	default:
		return fmt.Errorf("unknown command %q; see gh stars --help", command)
	}

	client, err := NewClient()
	if err != nil {
		return fmt.Errorf("initialize GitHub client (authenticate with gh auth login): %w", err)
	}
	switch command {
	case "lists":
		return client.Lists(ctx, out)
	case "list":
		return client.List(ctx, args[1], out)
	default:
		return client.All(ctx, out)
	}
}
