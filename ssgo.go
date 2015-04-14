package ssgo

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, credentials *Credentials)

type SSO interface {
	RedirectToLogin(w http.ResponseWriter, r *http.Request)
	ExchangeCodeForToken(w http.ResponseWriter, r *http.Request)
	LookupToken(r *http.Request) *Credentials
	Handle(handler AuthenticatedHandler) http.HandlerFunc
	Route(loggedOut http.HandlerFunc, loggedIn AuthenticatedHandler) http.HandlerFunc
}

type sso struct {
	conf     *oauth2.Config
	states   map[string]time.Time
	site     string
	authOpts []oauth2.AuthCodeOption
}

type Credentials struct {
	Site   string
	Token  *oauth2.Token
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

func (s *sso) LookupToken(r *http.Request) *Credentials {
	cookie, err := r.Cookie(s.cookieName())
	if err != nil {
		return nil
	}
	tok := oauth2.Token{}
	err = LookupBoltJson(s.bucketName(), cookie.Value, &tok)
	if err != nil {
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
