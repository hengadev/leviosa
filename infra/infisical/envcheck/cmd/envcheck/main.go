// Command envcheck checks env files rendered from the Infisical Agent
// templates and lists their key names, never their values.
//
//	envcheck files <dev|staging|prod>        the files that environment renders
//	envcheck check <dev|staging|prod> <dir>  check <dir>/<file> for each of them
//
// infra/infisical/check-env.sh renders the files from Infisical into a
// private temporary directory and runs `envcheck check` on it.
package main

import (
	"fmt"
	"os"

	"github.com/Leviosa-care/leviosa/infra/infisical/envcheck"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	switch {
	case len(args) == 2 && args[0] == "files":
		names, err := envcheck.Files(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "envcheck:", err)
			return 2
		}
		for _, n := range names {
			fmt.Println(n)
		}
		return 0
	case len(args) == 3 && args[0] == "check":
		ok, err := envcheck.CheckDir(args[1], args[2], os.Stdout)
		if err != nil {
			fmt.Fprintln(os.Stderr, "envcheck:", err)
			return 2
		}
		if !ok {
			return 1
		}
		return 0
	default:
		fmt.Fprintln(os.Stderr, "usage: envcheck files <dev|staging|prod>\n       envcheck check <dev|staging|prod> <dir>")
		return 2
	}
}
