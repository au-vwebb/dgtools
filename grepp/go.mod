module github.com/DavidGamba/dgtools/grepp

go 1.13

require (
	github.com/DavidGamba/ffind v0.6.1
	github.com/DavidGamba/go-getoptions v0.23.0
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b

	// workaround for error: //go:linkname must refer to declared function or variable
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
)
