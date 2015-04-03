#ssgo

This package aims to make it easy to build web applications in Go that use external sites as their primary user account system.

## Currently supported sites:
- **github**

## Planned integrations:
- **reddit**
- **imgur**

# Super easy github integration:

    package main
    
    import (
    	"fmt"
    	"io"
    	"net/http"
    	"os"
    
    	"github.com/captncraig/ssgo"
    	"github.com/google/go-github/github"
    )
    
    var gh ssgo.GithubSSO
    
    func main() {
    	gh = ssgo.NewGithubSSO(os.Getenv("GH_CLIENT_ID"), os.Getenv("GH_CLIENT_SECRET"), "public_repo,write:repo_hook")
    	http.HandleFunc("/login", gh.RedirectToLogin)
    	http.HandleFunc("/ghauth", gh.ExchangeCodeForToken)
    	http.HandleFunc("/", gh.Route(loggedOut, loggedIn))
    	http.ListenAndServe(":5675", nil)
    }
    
    func loggedOut(w http.ResponseWriter, r *http.Request) {
    	w.Header().Add("Content-Type", "text/html")
    	io.WriteString(w, "<h1>Welcome</h1><p>This site does stuff. Please <a href='/login'>Login with github</a></p>")
    }
    
    func loggedIn(w http.ResponseWriter, r *http.Request, ghUser *ssgo.GithubUser) {
    	w.Header().Add("Content-Type", "text/html")
    	fmt.Fprintf(w, "<h2>Welcome, %s. <img src='%s' height=50 width=50></img></h2> Your repositories: <ul>", ghUser.Login, ghUser.AvatarUrl)
    	client := ssgo.GithubApiClient(ghUser.AccessToken)
    	ghClient := github.NewClient(client)
    	repos, _, err := ghClient.Repositories.List("", nil)
    	if err != nil {
    		fmt.Fprint(w, "<strong>Error fetching repositories</strong>")
    	}
    	for _, repo := range repos {
    		fmt.Fprintf(w, "<li><a href='%s'>%s</a></li>", *repo.HTMLURL, *repo.Name)
    	}
    
    }
