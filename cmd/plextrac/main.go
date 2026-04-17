// plextrac is the command-line interface to a Plextrac tenant.
//
// Commands:
//
//	plextrac flaws list | get | create | update | delete | import
//	plextrac clients list | get
//	plextrac reports list | get
//	plextrac assets list | get
//	plextrac writeups list | apply
//	plextrac exports start | get
//	plextrac auth login
//
// Credentials are read from environment variables only; see README.md.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "-v", "--version", "version":
		fmt.Printf("plextrac %s (%s, %s)\n", version, commit, date)
		return
	case "-h", "--help", "help":
		usage()
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "auth":
		err = runAuth(ctx, args)
	case "flaws":
		err = runFlaws(ctx, args)
	case "clients":
		err = runClients(ctx, args)
	case "reports":
		err = runReports(ctx, args)
	case "assets":
		err = runAssets(ctx, args)
	case "writeups":
		err = runWriteups(ctx, args)
	case "exports":
		err = runExports(ctx, args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `plextrac - Plextrac SDK + CLI

Usage:
  plextrac <command> [subcommand] [flags]

Commands:
  auth      login        Exchange credentials for a JWT
  flaws     list         List flaws in a report
            get          Fetch a single flaw
            create       Create a flaw from JSON on stdin
            update       Update a flaw from JSON on stdin
            delete       Delete a flaw
            import       Bulk upsert findings.json
  clients   list|get     Manage clients
  reports   list|get     Manage reports
  assets    list|get     Manage assets
  writeups  list|apply   Content library
  exports   start|get    PDF/DOCX exports

Env vars:
  PLEXTRAC_URL, PLEXTRAC_USERNAME, PLEXTRAC_PASSWORD, PLEXTRAC_MFA,
  PLEXTRAC_API_KEY, PLEXTRAC_TOKEN, PLEXTRAC_CLIENT_ID, PLEXTRAC_REPORT_ID

Run 'plextrac <command> --help' for command-specific options.
`)
}

// parseFlags returns a FlagSet pre-configured with -h for a subcommand,
// printing its own usage on -h.
func parseFlags(name string, args []string, register func(*flag.FlagSet)) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	register(fs)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return fs, nil
}
