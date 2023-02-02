# gopad

Notepad in Golang and using nMage and ImGUI.

## Developer Details

### Hide Console on Windows

Running the executable on windows opens a terminal as well. To hide that we build with: `go build -ldflags -H=windowsgui .`

### Program Icon

`*.syso` files are used by `go build` on Windows to add information to the executable that Windows can read, such as version and icon.
We use `github.com/tc-hib/go-winres` to generate these via the command: `go-winres simply --icon gopad-icon.ico --manifest gui`
