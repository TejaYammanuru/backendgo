package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type VerifiedUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("[JWTAuthMiddleware] Middleware start")

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Println("[JWTAuthMiddleware] Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}
		fmt.Println("[JWTAuthMiddleware] Authorization header received:", authHeader)

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			fmt.Println("[JWTAuthMiddleware] Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		fmt.Println("[JWTAuthMiddleware] Extracted token:", token)

		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://localhost:3001/verify_token", bytes.NewBuffer([]byte{}))
		if err != nil {
			fmt.Println("[JWTAuthMiddleware] Failed to create request to auth server:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to auth server"})
			c.Abort()
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		fmt.Println("[JWTAuthMiddleware] Sending request to auth server")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("[JWTAuthMiddleware] Failed to contact auth server:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to contact auth server"})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		fmt.Println("[JWTAuthMiddleware] Auth server responded with status:", resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[JWTAuthMiddleware] Failed to read auth server response:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read auth server response"})
			c.Abort()
			return
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Println("[JWTAuthMiddleware] Token verification failed, auth server response body:", string(body))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
			c.Abort()
			return
		}

		fmt.Println("[JWTAuthMiddleware] Auth server response body:", string(body))

		var user VerifiedUser
		if err := json.Unmarshal(body, &user); err != nil {
			fmt.Println("[JWTAuthMiddleware] Invalid response from auth server:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from auth server"})
			c.Abort()
			return
		}

		roleMap := map[string]int{
			"member":    0,
			"librarian": 1,
			"admin":     2,
		}

		userRole, ok := roleMap[user.Role]
		if !ok {
			fmt.Println("[JWTAuthMiddleware] Unknown role:", user.Role)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user role"})
			c.Abort()
			return
		}

		fmt.Printf("[JWTAuthMiddleware] Verified user: ID=%d, Name=%s, Email=%s, Role=%d\n", user.ID, user.Name, user.Email, userRole)

		c.Set("userID", user.ID)
		c.Set("userRole", userRole)
		c.Set("userName", user.Name)
		c.Set("userEmail", user.Email)

		fmt.Println("[JWTAuthMiddleware] Middleware success, passing to next handler")
		c.Next()
	}
}
