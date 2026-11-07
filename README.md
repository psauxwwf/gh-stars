# gh-stars

GitHub CLI extension for starred repositories and star lists, written in Go.
Prints one result per line directly to stdout, without a pager.

## Install

Authenticate with `gh auth login`, then install from a published release:

```sh
gh extension install psauxwwf/gh-stars
```

Release binaries currently support **Linux amd64**. No Go or jq is required
to run a published binary. Authentication uses the standard GitHub CLI settings,
including `GH_TOKEN`, `GITHUB_TOKEN`, and `GH_HOST`.

## Usage

```sh
gh stars                         # All starred repository URLs
gh stars all                     # Same as above
gh stars lists                   # Names of all your star lists
gh stars list '/sec/net/tunnel'   # Repository URLs from this exact list name
gh stars --help
```

Repository output uses full URLs, for example `https://github.com/cli/cli`.
Every command fetches all pages. List names are matched exactly, including
slashes, spaces, and case; they are not URL slugs. Missing or ambiguous list
names and API failures return a nonzero exit status with an error on stderr.

## Development

Requirements: Go (see `go.mod`), [Task](https://taskfile.dev), and GNU binutils
(`strip` and `objcopy`).

```sh
task check
task build
./gh-stars lists

# Install the Go binary in this checkout as a local extension.
gh extension install .
gh stars list '/sec/net/tunnel'
```

`task build` creates the ignored `./gh-stars` Go binary directly in the repository
root; there is no shell wrapper. Re-run it after changing Go source. Optional local environment values
can be placed in `build.env` (ignored by Git).

## Releases

```sh
task package                       # Build artifacts only; no publication
task release                       # Default tag: vMM.DD.HH
task release TAG=v1.0.0             # Explicit tag for a second release that hour
```

Before publishing, commit and push the source to `psauxwwf/gh-stars`. Release
creation targets the current commit and uploads both:

- `bin/gh-stars_linux-amd64` — raw static executable used by `gh extension install`;
- `bin/gh-stars_linux_amd64.tar.gz` — archive containing `gh-stars` for manual installation.

The tar.gz alone is not sufficient for GitHub CLI's binary extension installer.
Release publication is manual via `task release`; building never publishes.
