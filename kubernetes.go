package bob

import (
	"encoding/json"

	v1 "k8s.io/api/core/v1"
)

func generateConfigMap(config *Config) ([]byte, error) {
	var configMap v1.ConfigMap
	configMap.Data = config.Flags
	return json.Marshal(&configMap)
}
