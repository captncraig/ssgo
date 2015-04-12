package hub

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/captncraig/ssgo"
	ghApi "github.com/google/go-github/github"
)

type GithubAuthenticatedHandler func(w http.ResponseWriter, r *http.Request, user *GithubUser)

// GithubSSO manages all sign ins to github, as well as tracking cookies issued to users and the associated github access tokens.
type GithubSSO interface {
	// Initiate the login process by redirecting the user to the github sign-in and approve page.
	RedirectToLogin(w http.ResponseWriter, r *http.Request)
	// This should be linked to the callback url github has associated with your application.
	// This handler takes care of exchanging the code for an access token, and will drop a cookie before redirecting back to "/".
	ExchangeCodeForToken(w http.ResponseWriter, r *http.Request)
	// Lookup a user from an http request. Will return nil if no valid cookie found.
	LookupUser(r *http.Request) *GithubUser
	// Make a choice based on the incoming request. If user is already authenticated, the loggedin handler will execute.
	// Otherwise, the loggedOut handler will execute.
	Route(loggedOut http.HandlerFunc, loggedIn GithubAuthenticatedHandler) http.HandlerFunc
}

type githubSSO struct {
	clientId, clientSecret string
	requiredScopes         string
}

const ghAuthBucketName = "ghAuth"

func init() {
	if err := ssgo.EnsureBoltBucketExists(ghAuthBucketName); err != nil {
		panic(err)
	}
}

func NewGithubSSO(clientId, clientSecret, scopes string) GithubSSO {
	return &githubSSO{
		clientId:       clientId,
		clientSecret:   clientSecret,
		requiredScopes: scopes,
	}
}

var ghStates = map[string]time.Time{}

func (g *githubSSO) RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	state := ssgo.RandSeq(10)
	ghStates[state] = time.Now()
	url := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&scope=%s&state=%s", g.clientId, g.requiredScopes, state)
	http.Redirect(w, r, url, 302)
}

func (g *githubSSO) ExchangeCodeForToken(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if _, ok := ghStates[state]; state == "" || !ok {
		w.WriteHeader(401)
		io.WriteString(w, "Unknown state detected")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(400)
		io.WriteString(w, "No code provided")
		return
	}
	exchangeUrl := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", g.clientId, g.clientSecret, code)
	res, err := http.Post(exchangeUrl, "text/plain", nil)
	if err != nil || res.StatusCode != 200 {
		w.WriteHeader(502)
		fmt.Fprintf(w, "%s , %d", err, res.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil || strings.Contains(string(body), "error=") {
		w.WriteHeader(500)
		return
	}
	eq := strings.IndexRune(string(body), '=')
	amp := strings.IndexRune(string(body), '&')
	if eq == -1 || amp == -1 || amp < eq {
		w.WriteHeader(500)
		return
	}
	accessToken := string(body[eq+1 : amp])
	cookieVal := ssgo.RandSeq(25)

	gh := ghApi.NewClient(githubApiClient(accessToken))
	u, _, err := gh.Users.Get("")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	uname := *u.Login
	avatar := *u.AvatarURL
	g.storeGithubToken(cookieVal, accessToken, uname, avatar)
	http.SetCookie(w, &http.Cookie{Name: "ghAuthToken", Path: "/", Value: cookieVal, Expires: time.Now().Add(90 * 24 * time.Hour)})
	http.Redirect(w, r, "/", 302)
}

func (g *githubSSO) LookupUser(r *http.Request) *GithubUser {
	cookie, err := r.Cookie("ghAuthToken")
	if err != nil {
		return nil
	}
	user := &GithubUser{}
	err = ssgo.LookupBoltJson(ghAuthBucketName, cookie.Value, user)
	if err != nil {
		return nil
	}
	return user
}

func (g *githubSSO) storeGithubToken(cookie, token, username, avatar string) error {
	return ssgo.StoreBoltJson(ghAuthBucketName, cookie, &GithubUser{username, token, avatar})
}

func (g *githubSSO) Route(loggedOut http.HandlerFunc, loggedIn GithubAuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := g.LookupUser(r)
		if user == nil {
			loggedOut(w, r)
		} else {
			loggedIn(w, r, user)
		}
	}
}

// Basic information about a user.
type GithubUser struct {
	Login, AccessToken, AvatarUrl string
}

type ghClient struct {
	token string
}

func (g *ghClient) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", "token "+g.token)
	return http.DefaultTransport.RoundTrip(r)
}

// Creates an http.Client that can be used to make authenticated requests to the github api
func (u *GithubUser) GithubApiClient() *http.Client {
	return githubApiClient(u.AccessToken)
}

func githubApiClient(accessToken string) *http.Client {
	return &http.Client{Transport: &ghClient{accessToken}}
}
