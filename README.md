#ssgo

This package aims to make it easy to build web applications in Go that use external sites as their primary user account system.

## Currently supported sites:
- **github**

## Planned integrations:
- **reddit**
- **imgur**

# Super easy github integration:

Make an sso object:

    gh, _ = hub.NewGithubSSO(os.Getenv("GH_CLIENT_ID"), 
       os.Getenv("GH_CLIENT_SECRET"),
       "public_repo,write:repo_hook")`

Link the provided http handlers to whatever endpoint you want them to live at:

    http.HandleFunc("/login", gh.RedirectToLogin)
    http.HandleFunc("/ghauth", gh.ExchangeCodeForToken)
