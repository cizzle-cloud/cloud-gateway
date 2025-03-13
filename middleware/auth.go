package middleware

import (
	"api_gateway/config"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var client = &http.Client{
	Timeout: 5 * time.Second,
}

func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new request to the /auth/validate endpoint on the same API Gateway instance
		req, err := http.NewRequest("GET", cfg.Env.ValidateAuthURL, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create GET request to auth/validate: %v", err)})
			c.Abort()
			return
		}

		redirectURL := cfg.Env.RedirectUnauthorizedURL

		// Find cookie by name
		//TODO: name is 'Authorization'. Can in the same context have the same name from another provider?
		authCookie, err := c.Request.Cookie("Authorization")
		if err != nil {
			log.Printf("[ MIDDLEWARE ] Unauthorized user. Redirecting to %s at %s\n", redirectURL, time.Now())
			c.Redirect(http.StatusFound, redirectURL)
			c.Abort()
			return
		}

		req.AddCookie(authCookie)

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create GET request to auth/validate: %v", err)})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("[ MIDDLEWARE ] Unauthorized user. Redirecting to %s at %s\n", redirectURL, time.Now())
			c.Redirect(http.StatusFound, redirectURL)
			c.Abort()
			return
		}

		// Continue to the next handler
		c.Next()
	}
}
