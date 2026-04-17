// Command upsert-findings is a minimal SDK demo: it reads a findings.json
// and upserts each into a Plextrac report.
//
// Usage:
//
//	export PLEXTRAC_URL=... PLEXTRAC_USERNAME=... PLEXTRAC_PASSWORD=...
//	export PLEXTRAC_CLIENT_ID=1234 PLEXTRAC_REPORT_ID=5678
//	go run ./examples/upsert-findings ./findings.json
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cresco/plextrac-go/internal/importer"
	"github.com/cresco/plextrac-go/pkg/plextrac"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: upsert-findings <findings.json>")
	}
	c, err := plextrac.New(
		os.Getenv("PLEXTRAC_URL"),
		plextrac.WithPasswordAuth(
			os.Getenv("PLEXTRAC_USERNAME"),
			os.Getenv("PLEXTRAC_PASSWORD"),
			plextrac.EnvMFA{},
		),
		plextrac.WithRetry(3, time.Second),
		plextrac.WithRateLimit(10),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if err := importer.Run(ctx, importer.Config{
		Client:   c,
		Path:     os.Args[1],
		ClientID: os.Getenv("PLEXTRAC_CLIENT_ID"),
		ReportID: os.Getenv("PLEXTRAC_REPORT_ID"),
		Mode:     "upsert",
		Out:      os.Stdout,
	}); err != nil {
		log.Fatal(err)
	}
}
