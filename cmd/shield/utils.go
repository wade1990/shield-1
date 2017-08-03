package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/starkandwayne/goutils/ansi"
	"github.com/starkandwayne/shield/api"
)

var errCanceled = fmt.Errorf("Canceling... ")

func BoolString(tf bool) string {
	if tf {
		return "Y"
	}
	return "N"
}

func CurrentUser() string {
	return fmt.Sprintf("%s@%s", os.Getenv("USER"), os.Getenv("HOSTNAME"))
}

func PrettyJSON(raw string) string {
	tmpBuf := bytes.Buffer{}
	err := json.Indent(&tmpBuf, []byte(raw), "", "  ")
	if err != nil {
		DEBUG("json.Indent failed with %s", err)
		return raw
	}
	return tmpBuf.String()
}

func DEBUG(format string, args ...interface{}) {
	if debug {
		content := fmt.Sprintf(format, args...)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = "DEBUG> " + line
		}
		content = strings.Join(lines, "\n")
		fmt.Fprintf(os.Stderr, "%s\n", content)
	}
}

func OK(f string, l ...interface{}) {
	if *opts.Raw {
		RawJSON(map[string]string{"ok": fmt.Sprintf(f, l...)})
		return
	}
	ansi.Printf("@G{%s}\n", fmt.Sprintf(f, l...))
}

func MSG(f string, l ...interface{}) {
	if !*opts.Raw {
		ansi.Printf("\n@G{%s}\n", fmt.Sprintf(f, l...))
	}
}

func DisplayBackend(cfg *api.Config) {
	if cfg.BackendURI() == "" {
		ansi.Fprintf(os.Stderr, "No current SHIELD backend\n\n")
	} else {
		ansi.Fprintf(os.Stderr, "Using @G{%s} (%s) as SHIELD backend\n\n", cfg.BackendURI(), cfg.Backend)
	}
}
