# Transmission TUI Written in Golang using Bubbletea.
*This is currently a work in progress and a learning experience -- expect many iterations and changes*

# Features
1. Overview of torrent stats inspired by the official web gui
2. Individual stats view displaying info, peers, and file list
3. Reannounce torrent(s) TODO
4. Relocate torrent(s)
5. Rename torrent(s) TODO
6. Remove torrent(s) TODO
7. Delete torrent(s) TODO
7. Verify torrent(s) TODO

# Installation
This project has the following dependencies which can be installed with this command `go get ....`
1. github.com/charmbracelet/bubbles/textinput
2. github.com/charmbracelet/bubbletea
3. github.com/hekmon/transmissionrpc/v2
4. github.com/kelseyhightower/envconfig
5. gopkg.in/yaml.v3

# Usage
Place the binary `binary` in your path, adjust settings in `$XDG_HOME/settings.yml` if need be, and run the program

## Bindings

# Configuration
Configuration is done by modifying the `settings.yml` file located at `$XDG_HOME`

A Default configuration is provided and can be modified to adjust what columns are shown, server credentials,
and bookmarks(frequently used paths)

# Contributing
Help is welcome and appreciated! Feel free to submit PRs and I will review them as I can.
