package e2e

import "fmt"

func (r *Repository) GitAddAll() {
	r.cmd("git", "add", "-A")
}

func (r *Repository) GitCommit(msg string) {
	r.cmd("git", "commit", "-m", fmt.Sprintf("\"%s\"", msg))
}
