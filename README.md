#ssgo

This package aims to make it easy to build web applications in Go that use external sites as their primary user account system.

## Currently supported sites:
- **github**

## Planned integrations:
- **reddit**
- **imgur**

# Super easy github integration:

Make an sso object:

    cid, secret := os.Getenv("GH_CLIENT_ID"),os.Getenv("GH_CLIENT_SECRET")
    gh, err := hub.NewGithubSSO(cid,secret,"public_repo,write:repo_hook")

Link the provided http handlers to whatever endpoint you want them to live at:

    http.HandleFunc("/login", gh.RedirectToLogin)
    http.HandleFunc("/ghauth", gh.ExchangeCodeForToken)

Use the `Route` helper to direct traffic based on a user's cookie value:
    
    http.HandleFunc("/", gh.Route(loggedOut, loggedIn))
    
The appropriate handler will be invoked for requests, and if the user is logged in to github, you will receive a populated `GithubUser` struct to your loggedIn handler.
