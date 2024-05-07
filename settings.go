package tiga

import (
	"fmt"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
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

// var onceConfig sync.Once
// var config *Configuration = nil

func NewConfig(env string) *Configuration {
	// onceConfig.Do(func() {
	// 	config = &Configuration{
	// 		viper.New(),
	// 		env,
	// 	}
	// })
	config := &Configuration{
		viper.New(),
		env,
	}
	return config

}
func (c Configuration) GetValue(key string) (interface{}, error) {
	if !strings.Contains(key, c.env) {
		key = fmt.Sprintf("%s.%s", c.env, key)
	}
	value := c.Get(key)
	return value, nil
}
func (c Configuration) GetString(key string) string {
	if !strings.Contains(key, c.env)&&!strings.HasPrefix(key,"common") {
		key = fmt.Sprintf("%s.%s", c.env, key)
	}
	value := c.Get(key)
	return value.(string)
}
func (c Configuration) GetStrings(key string) []string {
	if !strings.Contains(key, c.env) {
		key = fmt.Sprintf("%s.%s", c.env, key)
	}
	values := c.GetStringSlice(key)
	return values
}
func (c Configuration) GetInt(key string) int {
	if !strings.Contains(key, c.env) {
		key = fmt.Sprintf("%s.%s", c.env, key)
	}
	value := c.Get(key)
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}

}
func (c Configuration) UnmarshalKey(key string, rawVal any, opts ...viper.DecoderConfigOption) error {
	if !strings.Contains(key, c.env) {
		key = fmt.Sprintf("%s.%s", c.env, key)
	}
	return c.Viper.UnmarshalKey(key, rawVal, opts...)
}
func (c Configuration) GetEnv() string {
	env, ok := os.LookupEnv("RUN_MODE")
	if !ok {
		env = c.env
	}
	return env
}
func (c Configuration) SetEnv(env string) {
	c.env = env
}
func (c Configuration) GetConfigByEnv(env string, key string) interface{} {
	return c.Get(fmt.Sprintf("%s.%s", env, key))
}
func (c Configuration) SetConfig(key string, val string, env string) {
	c.Set(fmt.Sprintf("%s.%s", env, key), val)
}
func (c Configuration) load(dir string) bool {
	c.AddConfigPath(dir)
	c.SetConfigName("settings")
	c.SetConfigType("yaml")
	readErr := c.ReadInConfig()
	if readErr != nil {
		panic(readErr)
	}
	err := c.Unmarshal(c)
	return err == nil
}
func InitSettings(env string, settingDir string) *Configuration {
	config := NewConfig(env)
	// Config.load(wd)
	config.load(settingDir)
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		config.load(settingDir)
	})
	return config

}
