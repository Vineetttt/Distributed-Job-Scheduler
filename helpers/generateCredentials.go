package helpers

import "log"

type ModuleCredentials struct {
	Username string
	Password string
}

func GenerateCredentials() (ModuleCredentials, error) {
	username, err := GenerateRandomHex(16)
	if err != nil {
		log.Println("Error generating username:", err)
		return ModuleCredentials{}, err
	}

	password, err := GenerateRandomHex(16)
	if err != nil {
		log.Println("Error generating password:", err)
		return ModuleCredentials{}, err
	}

	return ModuleCredentials{
		Username: username,
		Password: password,
	}, nil
}
