package main

import (
	"io"
	"net/http"
	"os"

	"github.com/captncraig/ssgo"
	"github.com/google/go-github/github"
)

var gh ssgo.SSO

func main() {
	gh = ssgo.NewGithub(os.Getenv("GH_CLIENT_ID"), os.Getenv("GH_CLIENT_SECRET"), "public_repo", "write:repo_hook")
	http.HandleFunc("/login", gh.RedirectToLogin)
	http.HandleFunc("/ghauth", gh.ExchangeCodeForToken)
	http.HandleFunc("/", gh.Route(loggedOut, loggedIn))
	http.ListenAndServe(":5675", nil)
}

func loggedOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	io.WriteString(w, "<h1>Welcome</h1><p>This site does stuff. Please <a href='/login'>Login with github</a></p>")
}

func loggedIn(w http.ResponseWriter, r *http.Request, cred *ssgo.Credentials) {
	user, _, err := github.NewClient(cred.Client).Users.Get("")
	if err != nil {
		io.WriteString(w, err.Error())
	}
	io.WriteString(w, user.String())
}
