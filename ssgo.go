package ssgo

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

//Like an http.HandleFunc, but accepts a credentials object as a third argument.
type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, credentials *Credentials)

//Core interface for working with a third-party website
type SSO interface {
	//Redirect the request to the provider's authorization page
	RedirectToLogin(w http.ResponseWriter, r *http.Request)
	//Handle callback from authorization page. This needs to be hosted at the url that is registered with the provider.
	ExchangeCodeForToken(w http.ResponseWriter, r *http.Request)
	//Lookup the credentials for a given request from the cookie. Will return nil if no valid cookie is found.
	LookupToken(r *http.Request) *Credentials
	//Basic http handler that looks up the token for you and provides credentials to your handler. Credentials may be nil.
	Handle(handler AuthenticatedHandler) http.HandlerFunc
	//Select a handler based on whether the user has a valid cookie or not.
	Route(loggedOut http.HandlerFunc, loggedIn AuthenticatedHandler) http.HandlerFunc
<<<<<<< HEAD

	ClearCookie(w http.ResponseWriter)
=======
>>>>>>> 8d92bc7a130adcb2784f5e6629349e7fed992ab3
}

type sso struct {
	conf     *oauth2.Config
	states   map[string]time.Time
	site     string
	authOpts []oauth2.AuthCodeOption
}

//Container for a user's oauth credentials
type Credentials struct {
	// Shortname of site they are authenticated with
	Site string
	// Oauth token for user.
	Token *oauth2.Token
	// Http client with oauth credentials ready to go.
	Client *http.Client
}

func NewGithub(clientId, clientSecret string, scopes ...string) SSO {
	return newSSO(clientId, clientSecret, "", "github", github.Endpoint, scopes, nil)
}

func NewReddit(clientId, clientSecret, redirectUri string, scopes ...string) SSO {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://www.reddit.com/api/v1/authorize",
		TokenURL: "https://www.reddit.com/api/v1/access_token",
	}
	return newSSO(clientId, clientSecret, redirectUri, "reddit", endpoint, scopes, []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("duration", "permanent")})
}

func newSSO(clientId, clientSecret, redirectUri, site string, endpoint oauth2.Endpoint, scopes []string, opts []oauth2.AuthCodeOption) SSO {
	conf := oauth2.Config{}
	conf.Endpoint = endpoint
	conf.ClientID = clientId
	conf.ClientSecret = clientSecret
	conf.Scopes = scopes
	if redirectUri != "" {
		conf.RedirectURL = redirectUri
	}
	s := &sso{
		conf:     &conf,
		states:   map[string]time.Time{},
		site:     site,
		authOpts: opts,
	}
	EnsureBoltBucketExists(s.bucketName())
	return s
}

func (s *sso) RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	state := randSeq(10)
	s.states[state] = time.Now()
	http.Redirect(w, r, s.conf.AuthCodeURL(state, s.authOpts...), 302)
}

func (s *sso) ExchangeCodeForToken(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		url := "/"
		if err != nil {
			url += "?ssoError=" + err.Error()
		}
		http.Redirect(w, r, url, 302)
	}()
	state := r.FormValue("state")
	if _, ok := s.states[state]; state == "" || !ok {
		err = fmt.Errorf("bad-state")
		return
	}
	code := r.FormValue("code")
	if code == "" {
		err = fmt.Errorf("no-code")
		return
	}
	tok, err := s.conf.Exchange(oauth2.NoContext, code)

	if err != nil {
		return
	}
	cookieVal := randSeq(25)
	err = StoreBoltJson(s.bucketName(), cookieVal, tok)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{Name: s.cookieName(), Value: cookieVal, Path: "/", Expires: time.Now().Add(90 * 24 * time.Hour)})
}

<<<<<<< HEAD
func (s *sso) ClearCookie(w http.ResponseWriter) {
	c := &http.Cookie{Name: s.cookieName(), Value: "", Path: "/", Expires: time.Now().Add(-1 * time.Hour), MaxAge: -1}
	http.SetCookie(w, c)
}

=======
>>>>>>> 8d92bc7a130adcb2784f5e6629349e7fed992ab3
func (s *sso) LookupToken(r *http.Request) *Credentials {
	cookie, err := r.Cookie(s.cookieName())
	if err != nil {
		return nil
	}
	tok := oauth2.Token{}
	err = LookupBoltJson(s.bucketName(), cookie.Value, &tok)
	if err != nil || tok.AccessToken == "" {
		return nil
	}

	return &Credentials{
		Site:   s.site,
		Token:  &tok,
		Client: s.conf.Client(oauth2.NoContext, &tok),
	}
}

func (s *sso) Handle(handler AuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := s.LookupToken(r)
		handler(w, r, tok)
	}
}
func (s *sso) Route(loggedOut http.HandlerFunc, loggedIn AuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := s.LookupToken(r)
		if tok == nil {
			loggedOut(w, r)
		} else {
			loggedIn(w, r, tok)
		}
	}
}

func (s *sso) bucketName() string {
	return s.site + "Tokens"
}
func (s *sso) cookieName() string {
	return s.site + "Tok"
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
