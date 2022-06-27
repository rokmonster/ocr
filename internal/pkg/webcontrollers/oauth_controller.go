package webcontrollers

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
	ID    string `json:"sub"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type oAuth2Controller struct {
	conf    *oauth2.Config
	store   sessions.Store
	enabled bool
}

func (ctrl *oAuth2Controller) authHandler(c *gin.Context) {
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

	session.Set("userdata", clientDetails)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func (ctrl *oAuth2Controller) GetLoginURL(state string) string {
	return ctrl.conf.AuthCodeURL(state, oauth2.ApprovalForce)
}

func NewOAuth2Controller(engine *gin.Engine, clientId, secret, domains string) *oAuth2Controller {
	gob.Register(OAuthClientInfo{})

	tlsDomains := strings.Split(domains, ",")

	ctrl := &oAuth2Controller{conf: nil, enabled: false}
	if len(tlsDomains) > 0 && len(clientId) > 0 && len(secret) > 0 {
		redirectUrl := fmt.Sprintf("https://%s/oauth", tlsDomains[0])
		logrus.Infof("Initiliazing OAuth2 with redirect url: %v", redirectUrl)

		ctrl = &oAuth2Controller{
			store: memstore.NewStore([]byte("Gisooshei6eitiQu2coe7ohze2phuuQu")),
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

		engine.Use(sessions.Sessions("session_id", ctrl.store))
		engine.GET("/oauth", ctrl.authHandler)
	} else {
		logrus.Warn("No OAuth2 setup found")
	}

	return ctrl
}

func (ctrl *oAuth2Controller) Middleware() func(ctx *gin.Context) {
	if ctrl.enabled {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}

	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		data := session.Get("userdata").(*OAuthClientInfo)
		if data == nil {
			state := randToken()
			session.Set("state", state)
			session.Save()
			ctx.Redirect(http.StatusFound, ctrl.GetLoginURL(state))
		} else {
			ctx.Set("userdata", data)
			ctx.Next()
		}
	}
}
