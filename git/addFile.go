package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func main() {
	path := filepath.FromSlash("/Users/Dipesh/OneDrive/Useful/Studying/GO/cli/git/tmp")
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "dipsy88",
			Password: "Dipesh77",
		},
		URL:      "https://github.com/Dipsy88/web_scrapper.git",
		Progress: os.Stdout,
	})
	checkIfError(err)

	w, err := repo.Worktree()
	checkIfError(err)

	filename := filepath.Join(path, "ExampleFile")
	err = ioutil.WriteFile(filename, []byte("hello world {}"), 0644)
	checkIfError(err)

	// Add the file to the staging area
	_, err = w.Add("ExampleFile")
	checkIfError(err)

	// Verify the current status of the worktree using the method Status
	status, err := w.Status()
	checkIfError(err)

	fmt.Println(status)

	// Check commit
	commit, err := w.Commit("ExampleFile", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Dipesh Pradhan",
			Email: "dipesh@gmail.com",
			When:  time.Now(),
		},
	})
	checkIfError(err)

	obj, err := repo.CommitObject(commit)
	checkIfError(err)

	fmt.Println(obj)

	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "dipsy88",
			Password: "Dipesh77",
		},
	})
	checkIfError(err)

}

func checkIfError(e error) {
	if e != nil {
		panic(e)
	}
}
