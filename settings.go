package tiga

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
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
	Env string
}

var onceConfig sync.Once
var config *Configuration = nil

func NewConfig(env string) *Configuration {
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
		env = c.Env
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
func InitSettings(env string) Configuration {
	config := NewConfig(env)
	// wd, _ := os.Getwd()
	var abPath string

	_, filename, _, ok := runtime.Caller(0)
	if ok {
		parentPath := filepath.Dir(filename)

		abPath = path.Dir(parentPath)

	}
	// Config.load(wd)
	config.load(abPath)
	return *config

}
