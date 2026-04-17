package main

import (
	"fmt"
	"os"

	"github.com/hazcod/plextrac-go/pkg/plextrac"
)

// newClient builds a SDK client from environment variables. Credentials are
// never taken from argv to keep them out of ps/history.
func newClient() (*plextrac.Client, error) {
	base := os.Getenv("PLEXTRAC_URL")
	if base == "" {
		return nil, fmt.Errorf("PLEXTRAC_URL is required")
	}

	var opt plextrac.Option
	switch {
	case os.Getenv("PLEXTRAC_TOKEN") != "":
		opt = plextrac.WithToken(os.Getenv("PLEXTRAC_TOKEN"))
	case os.Getenv("PLEXTRAC_API_KEY") != "":
		opt = plextrac.WithAPIKey(os.Getenv("PLEXTRAC_API_KEY"))
	case os.Getenv("PLEXTRAC_USERNAME") != "" && os.Getenv("PLEXTRAC_PASSWORD") != "":
		opt = plextrac.WithPasswordAuth(
			os.Getenv("PLEXTRAC_USERNAME"),
			os.Getenv("PLEXTRAC_PASSWORD"),
			plextrac.EnvMFA{},
		)
	default:
		return nil, fmt.Errorf("set PLEXTRAC_TOKEN, PLEXTRAC_API_KEY, or PLEXTRAC_USERNAME + PLEXTRAC_PASSWORD")
	}
	ua := fmt.Sprintf("plextrac-cli/%s", version)
	return plextrac.New(base, opt, plextrac.WithUserAgent(ua))
}

// ids returns the client/report IDs from flags or environment.
func ids(clientFlag, reportFlag string) (string, string, error) {
	cid := clientFlag
	if cid == "" {
		cid = os.Getenv("PLEXTRAC_CLIENT_ID")
	}
	rid := reportFlag
	if rid == "" {
		rid = os.Getenv("PLEXTRAC_REPORT_ID")
	}
	if cid == "" {
		return "", "", fmt.Errorf("--client or PLEXTRAC_CLIENT_ID required")
	}
	return cid, rid, nil
}
