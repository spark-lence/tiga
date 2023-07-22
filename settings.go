package tiga

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type Settings interface {
	// GetValue 获取指定的参数值
	GetValue(key string) (interface{}, error)
}

type Configuration struct {
	// Log *Logger `ymal:"log"`
	*viper.Viper
	env string
}

var onceConfig sync.Once
var config *Configuration = nil

func newConfig(env string) *Configuration {
	onceConfig.Do(func() {
		config = &Configuration{
			viper.New(),
			env,
		}
	})
	return config

}
func (c Configuration) GetValue(key string) (interface{}, error) {
	value := c.Get(key)
	return value, nil
}
func (c Configuration) GetEnv() string {
	env, ok := os.LookupEnv("RUN_MODE")
	if !ok {
		env = c.env
	}
	return env
}
func (c Configuration) GetConfigByEnv(env string, key string) interface{} {
	return c.Get(fmt.Sprintf("%s.%s", env, key))
}
func (c Configuration) load(dir string) bool {
	c.AddConfigPath(dir)
	c.SetConfigName("settings")
	c.SetConfigType("yaml")
	readErr := c.ReadInConfig()
	if readErr != nil {
		return false
	}
	err := c.Unmarshal(c)
	return err == nil
}
func InitSettings(env string, settingDir string) *Configuration {
	config := newConfig(env)
	// Config.load(wd)
	config.load(settingDir)
	return config

}
