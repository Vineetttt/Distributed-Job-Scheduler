package services

import (
	"fmt"
	"strings"

	redishelper "bitbucket.org/fastbanking/ring-jobscheduler-service/redis_helper"
)

type ModuleService struct {
	Cache *redishelper.RedisCache
}

func (m ModuleService) AddModule(moduleName string, credentials string) error {
	err := m.Cache.HSet("modules", moduleName, credentials)
	if err != nil {
		return err
	}
	return nil
}

func (m ModuleService) GetModuleCredentials(moduleName string) (string, string, error) {
	value, err := m.Cache.HGet("modules", moduleName)
	if err != nil {
		return "", "", err
	}

	splitValues := strings.Split(value, ":")
	if len(splitValues) != 2 {
		return "", "", fmt.Errorf("invalid value")
	}

	username := splitValues[0]
	hashedPassword := splitValues[1]

	return username, hashedPassword, nil

}

func (m ModuleService) Exist(moduleName string) bool {
	exists, err := m.Cache.HExist("modules", moduleName)
	if err != nil {
		fmt.Println(err)
	}
	return exists
}
