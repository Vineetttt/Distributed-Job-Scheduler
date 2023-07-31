package auth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/helpers"
	redishelper "bitbucket.org/fastbanking/ring-jobscheduler-service/redis_helper"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/services"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	Cache redishelper.RedisCache
}

func (m Middleware) AuthMiddleware(c *gin.Context) {

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response := helpers.PrepareErrorResponse("authorization required", http.StatusUnauthorized)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	credentials := strings.SplitN(authHeader, " ", 2)
	if len(credentials) != 2 || credentials[0] != "Basic" {
		response := helpers.PrepareErrorResponse("invalid authorization header", http.StatusBadRequest)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	decodedCredentials, err := base64.StdEncoding.DecodeString(credentials[1])
	if err != nil {
		response := helpers.PrepareErrorResponse("failed to decode authorization header", http.StatusBadRequest)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	userCredentials := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(userCredentials) != 2 {
		response := helpers.PrepareErrorResponse("invalid authorization credentials", http.StatusBadRequest)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	moduleName := userCredentials[0]
	password := userCredentials[1]

	moduleService := &services.ModuleService{
		Cache: &m.Cache,
	}
	_, dbPassword, err := moduleService.GetModuleCredentials(moduleName)
	if err != nil {
		response := helpers.PrepareErrorResponse("invalid authorization credentials", http.StatusBadRequest)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	match := helpers.ComparePassword(dbPassword, password)
	if match {
		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusForbidden, helpers.PrepareErrorResponse("authentication failed", http.StatusForbidden))
		return
	}
}
