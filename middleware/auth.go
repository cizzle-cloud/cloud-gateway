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

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new request to the /auth/validate endpoint on the same API Gateway instance
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/validate", cfg.AuthURL), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create GET request to auth/validate: %v", err)})
			c.Abort()
			return
		}

		authPagePath := fmt.Sprintf("%s/%s", cfg.ApiGatewayURL, cfg.AuthPagePath)

		// Find cookie by name
		//TODO: name is 'Authorization'. Can in the same context have the same name from another provider?
		authCookie, err := c.Request.Cookie("Authorization")
		if err != nil {
			log.Printf("[ MIDDLEWARE ] Unauthorized user. Redirecting to %s at %s\n", authPagePath, time.Now())
			c.Redirect(http.StatusFound, authPagePath)
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
			log.Printf("[ MIDDLEWARE ] Unauthorized user. Redirecting to %s at %s\n", authPagePath, time.Now())
			c.Redirect(http.StatusFound, authPagePath)
			c.Abort()
			return
		}

		// Continue to the next handler
		c.Next()
	}
}
