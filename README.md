#ssgo

This package aims to make it easy to build web applications in Go that use external sites as their primary user account system.

## Currently supported sites:
- **github**

## Planned integrations:
- **reddit**
- **imgur**

# Super easy github integration:

`import "github.com/captncraig/ssgo/hub"`

Make an sso object:

    cid, secret := os.Getenv("GH_CLIENT_ID"),os.Getenv("GH_CLIENT_SECRET")
    gh := hub.NewGithubSSO(cid,secret,"public_repo,write:repo_hook")

Link the provided http handlers to whatever endpoint you want them to live at:

    http.HandleFunc("/login", gh.RedirectToLogin)
    http.HandleFunc("/ghauth", gh.ExchangeCodeForToken)

Use the `Route` helper to direct traffic based on a user's cookie value:
    
    http.HandleFunc("/", gh.Route(loggedOut, loggedIn))
    
The appropriate handler will be invoked for requests, and if the user is logged in to github, you will receive a populated `GithubUser` struct to your loggedIn handler.

`user.GithubApiClient()` will give you an `http.Client` that you can use with [go-github](https://github.com/google/go-github) to make authenticated requests for that user.

See [this example](https://github.com/captncraig/ssgo/blob/master/examples/github/main.go) for full working code.

## how it works:
Internally we store a randomly generated `authToken` cookie in the browser, which is a key into a boltDb database that stores the accessToken and some basic account info. You can control the db file name with the `ssgo.boltdb` environment variable if you so choose.

If your application wants to use the same bolt db as the sso system, you can use the helpers in the ssgo package to load or store json to your own bucket.


