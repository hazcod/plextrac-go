package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cresco/plextrac-go/internal/importer"
	"github.com/cresco/plextrac-go/pkg/plextrac"
)

func runFlaws(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plextrac flaws <list|get|create|update|delete|import>")
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "list":
		return flawsList(ctx, rest)
	case "get":
		return flawsGet(ctx, rest)
	case "create":
		return flawsCreate(ctx, rest)
	case "update":
		return flawsUpdate(ctx, rest)
	case "delete":
		return flawsDelete(ctx, rest)
	case "import":
		return flawsImport(ctx, rest)
	default:
		return fmt.Errorf("unknown flaws subcommand: %s", sub)
	}
}

func flawsList(ctx context.Context, args []string) error {
	var cid, rid string
	var per int
	_, err := parseFlags("flaws list", args, func(fs *flag.FlagSet) {
		fs.StringVar(&cid, "client", "", "client ID")
		fs.StringVar(&rid, "report", "", "report ID")
		fs.IntVar(&per, "per-page", 100, "page size")
	})
	if err != nil {
		return err
	}
	cid, rid, err = ids(cid, rid)
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	iter := c.Flaws.List(ctx, cid, rid, plextrac.ListOpts{PerPage: per})
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	for iter.Next(ctx) {
		if err := enc.Encode(iter.Value()); err != nil {
			return err
		}
	}
	return iter.Err()
}

func flawsGet(ctx context.Context, args []string) error {
	var cid, rid string
	fs, err := parseFlags("flaws get", args, func(fs *flag.FlagSet) {
		fs.StringVar(&cid, "client", "", "client ID")
		fs.StringVar(&rid, "report", "", "report ID")
	})
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: plextrac flaws get [--client ID] [--report ID] <flaw-id>")
	}
	cid, rid, err = ids(cid, rid)
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	f, err := c.Flaws.Get(ctx, cid, rid, fs.Arg(0))
	if err != nil {
		return err
	}
	return writeJSON(os.Stdout, f)
}

func flawsCreate(ctx context.Context, args []string) error {
	var cid, rid string
	_, err := parseFlags("flaws create", args, func(fs *flag.FlagSet) {
		fs.StringVar(&cid, "client", "", "client ID")
		fs.StringVar(&rid, "report", "", "report ID")
	})
	if err != nil {
		return err
	}
	cid, rid, err = ids(cid, rid)
	if err != nil {
		return err
	}
	var flaw plextrac.Flaw
	if err := readJSON(os.Stdin, &flaw); err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	out, err := c.Flaws.Create(ctx, cid, rid, &flaw)
	if err != nil {
		return err
	}
	return writeJSON(os.Stdout, out)
}

func flawsUpdate(ctx context.Context, args []string) error {
	var cid, rid string
	fs, err := parseFlags("flaws update", args, func(fs *flag.FlagSet) {
		fs.StringVar(&cid, "client", "", "client ID")
		fs.StringVar(&rid, "report", "", "report ID")
	})
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: plextrac flaws update <flaw-id>")
	}
	cid, rid, err = ids(cid, rid)
	if err != nil {
		return err
	}
	var flaw plextrac.Flaw
	if err := readJSON(os.Stdin, &flaw); err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	out, err := c.Flaws.Update(ctx, cid, rid, fs.Arg(0), &flaw)
	if err != nil {
		return err
	}
	return writeJSON(os.Stdout, out)
}

func flawsDelete(ctx context.Context, args []string) error {
	var cid, rid string
	fs, err := parseFlags("flaws delete", args, func(fs *flag.FlagSet) {
		fs.StringVar(&cid, "client", "", "client ID")
		fs.StringVar(&rid, "report", "", "report ID")
	})
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: plextrac flaws delete <flaw-id>")
	}
	cid, rid, err = ids(cid, rid)
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	return c.Flaws.Delete(ctx, cid, rid, fs.Arg(0))
}

func flawsImport(ctx context.Context, args []string) error {
	var cid, rid, mode, only, mapFile string
	var dryRun, minimal bool
	var workers int
	fs, err := parseFlags("flaws import", args, func(fs *flag.FlagSet) {
		fs.StringVar(&cid, "client", "", "client ID")
		fs.StringVar(&rid, "report", "", "report ID")
		fs.StringVar(&mode, "mode", "upsert", "upsert | create | update")
		fs.StringVar(&only, "only", "", "comma-separated finding IDs")
		fs.StringVar(&mapFile, "id-map", "", "audit_id -> flaw_id JSON (default: fetch)")
		fs.BoolVar(&dryRun, "dry-run", false, "print payloads, skip upload")
		fs.BoolVar(&minimal, "minimal", false, "drop fields.scores from payload")
		fs.IntVar(&workers, "workers", 0, "parallel workers (0 = client default)")
	})
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: plextrac flaws import [flags] <findings.json>")
	}
	cid, rid, err = ids(cid, rid)
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	return importer.Run(ctx, importer.Config{
		Client:   c,
		Path:     fs.Arg(0),
		ClientID: cid,
		ReportID: rid,
		Mode:     mode,
		Only:     only,
		IDMap:    mapFile,
		DryRun:   dryRun,
		Minimal:  minimal,
		Workers:  workers,
		Out:      os.Stdout,
	})
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func readJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	return dec.Decode(v)
}
