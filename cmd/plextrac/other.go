package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hazcod/plextrac-go/pkg/plextrac"
)

// --- auth ---

func runAuth(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "login" {
		return fmt.Errorf("usage: plextrac auth login")
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	// Trigger auth by issuing a harmless call that requires a token.
	if _, err := c.Clients.List(ctx, plextrac.ListOpts{PerPage: 1}).All(ctx); err != nil {
		return err
	}
	fmt.Println("authenticated")
	return nil
}

// --- clients ---

func runClients(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plextrac clients <list|get>")
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		iter := c.Clients.List(ctx, plextrac.ListOpts{})
		for iter.Next(ctx) {
			if err := writeJSON(os.Stdout, iter.Value()); err != nil {
				return err
			}
		}
		return iter.Err()
	case "get":
		fs, err := parseFlags("clients get", args[1:], func(_ *flag.FlagSet) {})
		if err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: plextrac clients get <client-id>")
		}
		out, err := c.Clients.Get(ctx, fs.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(os.Stdout, out)
	}
	return fmt.Errorf("unknown clients subcommand: %s", args[0])
}

// --- reports ---

func runReports(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plextrac reports <list|get>")
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		var cid string
		_, err := parseFlags("reports list", args[1:], func(fs *flag.FlagSet) {
			fs.StringVar(&cid, "client", "", "client ID")
		})
		if err != nil {
			return err
		}
		cid, _, err = ids(cid, "")
		if err != nil {
			return err
		}
		iter := c.Reports.List(ctx, cid, plextrac.ListOpts{})
		for iter.Next(ctx) {
			if err := writeJSON(os.Stdout, iter.Value()); err != nil {
				return err
			}
		}
		return iter.Err()
	case "get":
		var cid string
		fs, err := parseFlags("reports get", args[1:], func(fs *flag.FlagSet) {
			fs.StringVar(&cid, "client", "", "client ID")
		})
		if err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: plextrac reports get [--client ID] <report-id>")
		}
		cid, _, err = ids(cid, "")
		if err != nil {
			return err
		}
		out, err := c.Reports.Get(ctx, cid, fs.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(os.Stdout, out)
	}
	return fmt.Errorf("unknown reports subcommand: %s", args[0])
}

// --- assets ---

func runAssets(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plextrac assets <list|get>")
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		var cid string
		_, err := parseFlags("assets list", args[1:], func(fs *flag.FlagSet) {
			fs.StringVar(&cid, "client", "", "client ID")
		})
		if err != nil {
			return err
		}
		cid, _, err = ids(cid, "")
		if err != nil {
			return err
		}
		iter := c.Assets.List(ctx, cid, plextrac.ListOpts{})
		for iter.Next(ctx) {
			if err := writeJSON(os.Stdout, iter.Value()); err != nil {
				return err
			}
		}
		return iter.Err()
	case "get":
		var cid string
		fs, err := parseFlags("assets get", args[1:], func(fs *flag.FlagSet) {
			fs.StringVar(&cid, "client", "", "client ID")
		})
		if err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: plextrac assets get [--client ID] <asset-id>")
		}
		cid, _, err = ids(cid, "")
		if err != nil {
			return err
		}
		out, err := c.Assets.Get(ctx, cid, fs.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(os.Stdout, out)
	}
	return fmt.Errorf("unknown assets subcommand: %s", args[0])
}

// --- writeups ---

func runWriteups(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plextrac writeups <list|apply>")
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		iter := c.Writeups.List(ctx, plextrac.ListOpts{})
		for iter.Next(ctx) {
			if err := writeJSON(os.Stdout, iter.Value()); err != nil {
				return err
			}
		}
		return iter.Err()
	case "apply":
		var cid, rid string
		fs, err := parseFlags("writeups apply", args[1:], func(fs *flag.FlagSet) {
			fs.StringVar(&cid, "client", "", "client ID")
			fs.StringVar(&rid, "report", "", "report ID")
		})
		if err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: plextrac writeups apply [--client ID] [--report ID] <writeup-id>")
		}
		cid, rid, err = ids(cid, rid)
		if err != nil {
			return err
		}
		out, err := c.Writeups.Apply(ctx, cid, rid, fs.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(os.Stdout, out)
	}
	return fmt.Errorf("unknown writeups subcommand: %s", args[0])
}

// --- exports ---

func runExports(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plextrac exports <start|get>")
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	switch args[0] {
	case "start":
		var cid, rid, format string
		_, err := parseFlags("exports start", args[1:], func(fs *flag.FlagSet) {
			fs.StringVar(&cid, "client", "", "client ID")
			fs.StringVar(&rid, "report", "", "report ID")
			fs.StringVar(&format, "format", "pdf", "export format")
		})
		if err != nil {
			return err
		}
		cid, rid, err = ids(cid, rid)
		if err != nil {
			return err
		}
		out, err := c.Exports.Start(ctx, cid, rid, format)
		if err != nil {
			return err
		}
		return writeJSON(os.Stdout, out)
	case "get":
		var cid, rid string
		fs, err := parseFlags("exports get", args[1:], func(fs *flag.FlagSet) {
			fs.StringVar(&cid, "client", "", "client ID")
			fs.StringVar(&rid, "report", "", "report ID")
		})
		if err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: plextrac exports get <export-id>")
		}
		cid, rid, err = ids(cid, rid)
		if err != nil {
			return err
		}
		out, err := c.Exports.Get(ctx, cid, rid, fs.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(os.Stdout, out)
	}
	return fmt.Errorf("unknown exports subcommand: %s", args[0])
}
