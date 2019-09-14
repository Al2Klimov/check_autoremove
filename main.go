//go:generate go run vendor/github.com/Al2Klimov/go-gen-source-repos/main.go github.com/Al2Klimov/check_autoremove

package main

import (
	"bytes"
	"flag"
	"fmt"
	_ "github.com/Al2Klimov/go-gen-source-repos"
	. "github.com/Al2Klimov/go-monplug-utils"
	"html"
	"os"
	"os/exec"
	"strings"
)

func main() {
	os.Exit(ExecuteCheck(onTerminal, checkAutoremove))
}

func onTerminal() (output string) {
	return fmt.Sprintf(
		"For the terms of use, the source code and the authors\n"+
			"see the projects this program is assembled from:\n\n  %s\n",
		strings.Join(GithubcomAl2klimovGo_gen_source_repos, "\n  "),
	)
}

func checkAutoremove() (output string, perfdata PerfdataCollection, errs map[string]error) {
	cli := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var warn, crit OptionalThreshold

	cli.Var(&warn, "warn", "e.g. @~:42")
	cli.Var(&crit, "crit", "e.g. @~:42")

	if errCli := cli.Parse(os.Args[1:]); errCli != nil {
		os.Exit(3)
	}

	cmd := exec.Command("apt-get", "-sqq", "autoremove")
	var out bytes.Buffer

	cmd.Env = []string{"LC_ALL=C"}
	cmd.Dir = "/"
	cmd.Stdin = nil
	cmd.Stdout = &out
	cmd.Stderr = nil

	if errRun := cmd.Run(); errRun != nil {
		errs = map[string]error{"apt-get -sqq autoremove": errRun}
		return
	}

	lines := bytes.Split(out.Bytes(), []byte{'\n'})
	if len(lines) > 0 && len(lines[len(lines)-1]) < 1 {
		lines = lines[:len(lines)-1]
	}

	perfdata = PerfdataCollection{Perfdata{
		Label: "packages",
		Value: float64(len(lines)),
		Warn:  warn,
		Crit:  crit,
		Min:   OptionalNumber{true, 0},
	}}

	if len(lines) > 0 {
		var buf bytes.Buffer
		prefix := []byte("Remv ")

		buf.Write([]byte("<p><b>Some packages may be removed</b></p>\n\n<ul>"))

		for _, line := range lines {
			buf.Write([]byte("<li>"))
			buf.Write([]byte(html.EscapeString(string(bytes.TrimPrefix(line, prefix)))))
			buf.Write([]byte("</li>"))
		}

		buf.Write([]byte("</ul>"))

		output = buf.String()
	} else {
		output = "<p>No packages to remove.</p>"
	}

	return
}
