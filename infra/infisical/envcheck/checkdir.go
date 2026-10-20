package envcheck

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CheckDir checks every file env renders, read from dir (one file per name,
// as check-env.sh writes them), and prints each file's key names, problems
// and warnings to w — never a value. It returns whether no file has a
// problem.
func CheckDir(env, dir string, w io.Writer) (bool, error) {
	names, err := Files(env)
	if err != nil {
		return false, err
	}
	fmt.Fprintf(w, "envcheck: Infisical %s, %d files (key names only, never values)\n", env, len(names))
	total := 0
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(dir, name))
		var r Report
		if err != nil {
			r = Report{File: name, Problems: []string{"not rendered"}}
		} else {
			r = Check(env, name, content)
		}
		fmt.Fprintf(w, "\n%s (%d keys)\n", name, len(r.Keys))
		for _, k := range r.Keys {
			fmt.Fprintf(w, "  %s\n", k)
		}
		for _, p := range r.Problems {
			fmt.Fprintf(w, "  PROBLEM: %s\n", p)
		}
		for _, wn := range r.Warnings {
			fmt.Fprintf(w, "  warning: %s\n", wn)
		}
		total += len(r.Problems)
	}
	if total > 0 {
		fmt.Fprintf(w, "\nenvcheck: %s: %d problem(s)\n", env, total)
		return false, nil
	}
	fmt.Fprintf(w, "\nenvcheck: %s: OK\n", env)
	return true, nil
}
