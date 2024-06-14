package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2api "google.golang.org/api/oauth2/v2"
)

var (
	googleOauthConfig *oauth2.Config
	oauthStateString  = "random" // Random string for security
)

func init() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),     // Set your Client ID
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"), // Set your Client Secret
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}
func main() {
	r := gin.Default()
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	r.GET("/", func(c *gin.Context) {
		data := struct {
			Title    string
			Message  string
			LoggedIn bool
		}{
			Title:   "Welcome",
			Message: "Please log in with Google",
		}
		if user, _ := c.Get("user"); user != nil {
			data.Message = "logged-in"
			data.LoggedIn = true
		}
		tmpl.Execute(c.Writer, data)
	})

	r.GET("/login", func(c *gin.Context) {
		url := googleOauthConfig.AuthCodeURL(oauthStateString)
		c.Redirect(http.StatusTemporaryRedirect, url)
	})

	r.GET("/auth/callback", handleGoogleCallback)

	r.Run(":8080")
}

func handleGoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	service, err := oauth2api.New(client)
	if err != nil {
		fmt.Printf("oauth2api.New() failed with '%s'\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	userinfo, err := service.Userinfo.Get().Do()
	if err != nil {
		fmt.Printf("Userinfo.Get() failed with '%s'\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// Save the user information in the context
	c.Set("user", userinfo)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
