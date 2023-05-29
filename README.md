# Transmission TUI Written in Golang using Bubbletea.
This is currently a work in progress and a learning experience -- expect many iterations and changes

# Features
1. Overview of torrent stats inspired by the official web gui
2. Individual stats view displaying info, peers, and file list
3. Reannounce
4. Relocate
5. Remove
6. Delete

# Installation
TBD....

# Usage
This project has the following dependencies which can be installed with this command `go get ....`
1. github.com/charmbracelet/bubbles/textinput
2. github.com/charmbracelet/bubbletea
3. github.com/hekmon/transmissionrpc/v2
4. github.com/kelseyhightower/envconfig
5. gopkg.in/yaml.v3

# Configuration
Configuration is done by modifying the `settings.yml` file located at `$XDG_HOME`


A Default configuration is provided and can be modified to adjust what columns are shown, server credentials,
and bookmarks(frequently used paths)

# Contributing
Help is welcome and appreciated! Feel free to submit PRs and I will review them as I can.
