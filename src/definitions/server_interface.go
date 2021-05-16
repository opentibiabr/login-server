package definitions

import "github.com/opentibiabr/login-server/src/configs"

type ServerInterface interface {
	Run(globalConfigs configs.GlobalConfigs) error
	GetName() string
}
