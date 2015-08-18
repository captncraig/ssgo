package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/captncraig/ssgo"
)

var sso = ssgo.NewReddit(os.Getenv("REDDIT_ID"), os.Getenv("REDDIT_SECRET"), os.Getenv("REDDIT_REDIRECT"), "identity")

func main() {
	http.HandleFunc("/login", sso.RedirectToLogin)
	http.HandleFunc("/redditAuth", sso.ExchangeCodeForToken)
	http.HandleFunc("/", sso.Route(loggedOut, loggedIn))
	http.ListenAndServe(":5675", nil)
}

func loggedOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	io.WriteString(w, "<h1>Welcome</h1><p>This site does stuff. Please <a href='/login'>Login with reddit</a></p>")
}

func loggedIn(w http.ResponseWriter, r *http.Request, c *ssgo.Credentials) {
	fmt.Println(c.Token.RefreshToken)
	resp, err := c.Client.Get("https://oauth.reddit.com/api/v1/me")
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, string(body))
}
