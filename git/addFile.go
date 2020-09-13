package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	pathInput := flag.String("path", "/Users/tmp/git", "Path to clone github repo")
	userName := flag.String("userName", "", "User name for the git repo")
	email := flag.String("email", "", "Email address associated with the user name for the git repo")
	repoURL := flag.String("repo", "", "Git repo to clone")
	flag.Parse()

	if *userName == "" || *repoURL == "" {
		flag.Usage()
		os.Exit(1)
	}
	fmt.Print("Enter password: ")
	password, _ := terminal.ReadPassword(int(syscall.Stdin))

	path := filepath.FromSlash(*pathInput)
	makeDirIfRequired(path)

	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: *userName,
			Password: string(password[:]),
		},
		URL:      *repoURL,
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
			Name:  *userName,
			Email: *email,
			When:  time.Now(),
		},
	})
	checkIfError(err)

	obj, err := repo.CommitObject(commit)
	checkIfError(err)
	fmt.Println(obj)

	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: *userName,
			Password: string(password[:]),
		},
	})
	checkIfError(err)
}

func makeDirIfRequired(path string) {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		fmt.Println("Creating directory", path)
		err := os.MkdirAll(path, 0755)
		checkIfError(err)
	}
}

func checkIfError(e error) {
	if e != nil {
		panic(e)
	}
}
