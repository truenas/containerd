package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type config struct {
	appsDataset string
}

func loadConfig() (map[string]interface{}, error) {
	var configPath = fmt.Sprintf("%s/%s", configDir, configFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]interface{})
	err = json.Unmarshal(data, &configMap)
	if err != nil {
		logrus.Errorf("Failed to load configuration for middleware: %s", err)
	}
	return configMap, err
}

func parseValue(name string, configMap map[string]interface{}, defaultValue bool) bool {
	value, ok := configMap[name]
	if ok {
		return value.(bool)
	}
	return defaultValue
}

func parseStringListValue(name string, configMap map[string]interface{}, defaultValue []string) []string {
	value, ok := configMap[name]
	if ok {
		var stringList []string
		for _, val := range value.([]interface{}) {
			strVal, ok := val.(string)
			if ok {
				stringList = append(stringList, strVal)
			}
		}
		return stringList
	}
	return defaultValue
}

func InitConfig() error {
	configMap, err := loadConfig()
	if err != nil {
		return err
	}
	requiredKeys := [1]string{"appsDataset"}
	for _, key := range requiredKeys {
		if _, ok := configMap[key]; !ok {
			errString := fmt.Sprintf("%s key must be specified", key)
			return errors.New(errString)
		}
	}

	clientConfig = &config{}
	clientConfig.appsDataset = configMap["appsDataset"].(string)
	return nil
}
