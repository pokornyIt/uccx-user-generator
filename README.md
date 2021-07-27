# UCCX User Generator

Tool for generate and delete multiple users on CISCO Contact Center Express (UCCX). Useful for testing any custom
synchronization tools cooperated with UCCX.

## Description

Tool manipulates with own users and teams in CISCO Contact Center Express. The tool allows you to add or delete own
users and associate them with own teams, so the final number would correspond to the required number. The tool creates a
team for every 10 own users.

Example users and teams:

- 100 users in 10 teams
- 500 users in 50 teams
- 2000 users in 200 teams
- 10000 users in 1000 teams

# Commands

Usage: uccx-user-generator -n X -x srv -u usr -p pwd -c srv -a usr -s pwd [options]

Flags:

* **-n X** - Number of expected users
* **-x srv** - UCCX server FQDN or IP address
* **-u usr** - UCCX administrator name
* **-p pwd** - UCCX administrator password
* **-c srv** - CUCM Publisher FQDN or IP address
* **-a usr** - CUCM AXL administrator name
* **-s pwd** - CUCM AXL administrator password

Options:

* **-t 30** - Request timeout in seconds (5 - 120)
* **-h** - Show context-sensitive help
* **-l INFO** - Logging level (Fatal, Error, Warning, Info, Debug, Trace)
* **--version** - Show application version.

Number **X** is expected between 0 and 10000. Set **X** to 0 is removed all generated users and teams.

# Sponsoring

This tool sponsored by [ZOOM International, a.s.](https://eleveo.com)   