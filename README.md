git-recent-changes
==================

This is a tool for getting a list of recently merged patches in a git
repository.  It assumes that the repository is using a bot to merge pull
requests once they've been approved.

The output is a list of pull request titles, summaries, and links.  The output
is designed to be email-friendly.  JSON output format is experimental.

Installation
------------

```
$ go install github.com/honza/git-recent-changes
```

Configuration
-------------

Under the hood, we access GitHub's API to get pull request data.  In order to
avoid hitting rate limits, you need to use a personal token.  The token doesn't
need any special privileges.

Create a file at `~/.config/gitrecentchanges.ini` with the following:

``` ini
username = <github username>
token = <github token>
```

Usage
-----

Clone the repository first, and `cd` into it.

There are at least two pieces of information we need:

1.  The GitHub repository for this project
2.  How far back we should go when collecting commits to show

### The GitHub repository for this project

If your project has a single origin, use the upstream option:

```
$ git-recent-changes --upstream-repo <github user>/<github repo>
```

If your project has a downstream, and an upstream origin, you can specify both:

```
$ git-recent-changes \
    --upstream-repo <github user>/<github repo>
    --downstream-repo <github user>/<github repo>
```

If both are specified, we will first look for a pull request upstream, and then
downstream.  The assumption is that downstream mostly tracks what upstream does,
and as such, any context for the change is more likely to be in the upstream
repository.

### How far back we should go when collecting commits to show

You can either pass in a git ref with `--end-ref <ref>`, or use the `--since` flag.

```
git-recent-changes

Usage:
  git-recent-changes [flags]

Flags:
  -c, --config string            Path to a config file (default "~/.config/gitrecentchanges.ini")
  -d, --dir string               Directory where the git repository is found (default ".")
      --downstream-repo string
      --end-ref string           Where should we stop? (default "HEAD~1")
  -h, --help                     help for git-recent-changes
  -i, --indent int               How many spaces to indent? (default 8)
  -o, --output-format string     options: plain, json (default "plain")
      --since string             e.g. 2021-03-21
      --start-ref string         Where should we start? (default "HEAD")
      --upstream-repo string
  -v, --verbose                  Show extra output
```

Example
-------

Running against [metal3-io/baremetal-operator](https://github.com/metal3-io/baremetal-operator) when `HEAD` is at `d3bcdd79e5114eef4901d82a73983090aa7679a6`:

```
git-recent-changes --upstream-repo metal3-io/baremetal-operator --end-ref HEAD~3
```

Will produce the following:

```
deploy.sh: use getopts to parse arguments

        PR: https://github.com/metal3-io/baremetal-operator/pull/1078


Update CI badges with Metal3

        PR: https://github.com/metal3-io/baremetal-operator/pull/1081

        Airship name from CI jenkins view have been deprecated


Move from Available to Preparing if HostFirmwareSettings changed

        PR: https://github.com/metal3-io/baremetal-operator/pull/1075

        Currently if the host is in the Available state and the HFS settings
        are changed it doesn't transition to the Preparing state as it would
        when the BareMetalHost settings are changed. This change fixes that.
```

License
-------

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
