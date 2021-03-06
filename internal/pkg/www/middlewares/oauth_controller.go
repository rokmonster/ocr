package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	AuthUserData = "rokmonster.dev/auth/userdata"
)

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

type OAuthCredentions struct {
	ClientID string `json:"id"`
	Secret   string `json:"secret"`
}

type OAuthClientInfo struct {
	ID      string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

type oauth2Middleware struct {
	conf    *oauth2.Config
	store   sessions.Store
	enabled bool
}

func (ctrl *oauth2Middleware) authHandler(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: %s", retrievedState))
		return
	}

	tok, err := ctrl.conf.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := ctrl.conf.Client(context.Background(), tok)
	userInfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer userInfo.Body.Close()
	data, _ := ioutil.ReadAll(userInfo.Body)

	var clientDetails OAuthClientInfo
	json.Unmarshal(data, &clientDetails)

	session.Set(AuthUserData, clientDetails)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func (ctrl *oauth2Middleware) GetLoginURL(state string) string {
	return ctrl.conf.AuthCodeURL(state, oauth2.ApprovalForce)
}

func NewOAuth2Middleware(engine *gin.Engine, clientId, secret, domains string) *oauth2Middleware {
	gob.Register(OAuthClientInfo{})

	sessionStore := memstore.NewStore([]byte("Gisooshei6eitiQu2coe7ohze2phuuQu"))
	engine.Use(sessions.Sessions("session_id", sessionStore))

	engine.GET("/logout", func(ctx *gin.Context) {
		sess := sessions.Default(ctx)
		sess.Clear()
		_ = sess.Save()

		ctx.Redirect(http.StatusFound, "/")
	})

	ctrl := &oauth2Middleware{conf: nil, enabled: false, store: sessionStore}

	tlsDomains := strings.Split(domains, ",")
	if len(tlsDomains) > 0 && len(clientId) > 0 && len(secret) > 0 {
		redirectUrl := fmt.Sprintf("https://%s/oauth", tlsDomains[0])
		logrus.Infof("Initiliazing OAuth2 with redirect url: %v", redirectUrl)

		ctrl = &oauth2Middleware{
			enabled: true,
			store:   sessionStore,
			conf: &oauth2.Config{
				ClientID:     clientId,
				ClientSecret: secret,
				RedirectURL:  redirectUrl,
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
				},
				Endpoint: google.Endpoint,
			},
		}

		engine.GET("/oauth", ctrl.authHandler)
		engine.GET("/login", ctrl.Login)
	} else {
		logrus.Warn("No OAuth2 setup found")
	}

	return ctrl
}

func (ctrl *oauth2Middleware) Login(ctx *gin.Context) {
	session := sessions.Default(ctx)

	state := randToken()
	session.Set("state", state)
	session.Save()
	ctx.Redirect(http.StatusFound, ctrl.GetLoginURL(state))
}

func (ctrl *oauth2Middleware) Middleware() func(ctx *gin.Context) {
	if !ctrl.enabled {
		return func(ctx *gin.Context) {
			ctx.Set(AuthUserData, OAuthClientInfo{})
			ctx.Next()
		}
	}

	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		data := session.Get(AuthUserData)
		if data == nil {
			ctx.HTML(200, "auth.html", gin.H{})
			ctx.Abort()
		} else {
			ctx.Set(AuthUserData, data)
			ctx.Next()
		}
	}
}
