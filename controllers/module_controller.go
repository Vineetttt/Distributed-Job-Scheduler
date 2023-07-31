package controllers

import (
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/helpers"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/services"
	"github.com/gin-gonic/gin"
)

var requestBody struct {
	ModuleName string `json:"moduleName"`
}

type ModuleController struct {
	ModuleService *services.ModuleService
}

func (m ModuleController) AddModule(c *gin.Context) {
	now := time.Now()
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		response := helpers.PrepareErrorResponse("invalid request body", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	moduleName := requestBody.ModuleName
	fmt.Println("1", time.Since(now))
	exists := m.ModuleService.Exist(moduleName)
	// if the module exists throw error
	if exists {
		response := helpers.PrepareErrorResponse("module already exists", http.StatusConflict)
		c.JSON(http.StatusConflict, response)
		return
	}
	fmt.Println("2", time.Since(now))
	// if it does not exist then generate new credentials
	moduleCred, err := helpers.GenerateCredentials()
	if err != nil {
		response := helpers.PrepareErrorResponse("failed to generate credentials", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	fmt.Println("3", time.Since(now))
	// password hashing
	hashedPass, err := helpers.HashPassword(moduleCred.Password)
	if err != nil {
		response := helpers.PrepareErrorResponse("failed to hash password", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	fmt.Println("4", time.Since(now))
	combinedValue := moduleCred.Username + ":" + hashedPass
	// storing the credentials in redis
	// key (moduleName) -> value (combinedValue)
	err = m.ModuleService.AddModule(moduleName, combinedValue)
	if err != nil {
		response := helpers.PrepareErrorResponse("failed to store credentials", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	fmt.Println("5", time.Since(now))
	// if everything goes well
	data := map[string]interface{}{
		"moduleName": moduleName,
		"username":   moduleCred.Username,
		"password":   moduleCred.Password,
	}
	response := helpers.PrepareSuccessResponse("success", http.StatusOK, data)
	c.JSON(http.StatusOK, response)
}
