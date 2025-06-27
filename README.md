# A Boot.dev project.

## Fetches RSS feeds from urls.

## Requirements
- `go 1.24.3`
- A working connection to a database is required. For example, I'm running *postrgesql* in docker on my computer with the default port and here's my connection string. You'll have to change it based on your setup.
`{"db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable","current_user_name":"<user>"}`
- This config file containing the connection string and username should be named `.gatorconfig` and stored at `~/.gatorconfig`

## Installation
Gator can be installed with `go install github.com/luckyhut/gator@latest`

## Basic usage
`gator` requires a user to be logged in. You can create a user with
`gator register <name>` where <name> is the username you'd like to use. Creating a user automatically logs you in, but `gator login` can be used to log in as an existing user.

Once you have a username, you can add sites you want RSS updates from.
`gator register "<site name>" "<url>" `
ex: `gator register "Boot Dev" "https://blog.boot.dev/index.xml"`

Gator is designed to be run from the terminal as a daemon. The `agg` command is designed to be used with an update interval to fetch after a given amount of time. 
`gator agg 10m` look for new posts every 10 minutes
`gator agg 4h` look for new posts every 4 hours
`gator agg 1d` look for new posts every day

Once you have some posts to read, use `browse` with an optional argument to list given RSS posts. 
`gator browse` displays 2 posts (default behavior)
`gator browse 5` displays 5 posts
