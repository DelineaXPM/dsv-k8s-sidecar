package main

// tools is a list of tools that are installed as binaries for development usage.
// This list gets installed to go bin directory once `mage init` is run.
// This is for binaries that need to be invoked as cli tools, not packages.
var toolList = []string{ //nolint:gochecknoglobals // ok to be global for tooling setup
}
