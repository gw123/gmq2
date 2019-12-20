// + !debug

package core

import (
	"github.com/gw123/gmq/core/interfaces"
	"github.com/spf13/viper"
	"os"
	"regexp"
)

/*
 * 模块管理模块
 * 1. 加载模块,卸载模块
 * 2. 管理配置,更新模块配置
 */
type ConfigManager struct {
	ModuleConfigs map[string]interfaces.ModuleConfig
	app           interfaces.App
	GlobalConfig  interfaces.AppConfig
	ConfigData    *viper.Viper
}

func NewConfigManager(app interfaces.App, configData *viper.Viper) *ConfigManager {
	this := new(ConfigManager)
	this.app = app
	this.ModuleConfigs = make(map[string]interfaces.ModuleConfig)
	this.GlobalConfig = NewAppConfig()
	this.ConfigData = configData
	err := this.ParseConfig()
	if err != nil {
		this.app.Warn("ConfigManger", "配置文件解析失败 "+err.Error())
	}
	return this
}

func (this *ConfigManager) ParseConfig() (err error) {
	globalConfig := this.ConfigData.GetStringMapString("app")
	reg := regexp.MustCompile(`^\$\{(.*)\}$`)
	for key, val := range globalConfig {
		if reg.MatchString(val) {
			arrs := reg.FindStringSubmatch(val)
			if len(arrs) > 1 {
				this.app.Debug("app", "读取环境变量 %s", arrs[1])
				val = os.Getenv(arrs[1])
			}
		}
		this.GlobalConfig.SetItem(key, val)
	}

	modulesConfig := this.ConfigData.GetStringMap("modules")
	for moduleName, moduleConfig := range modulesConfig {
		configs, ok := moduleConfig.(map[string]interface{})
		moduleConfig := NewModuleConfig(moduleName, this.GlobalConfig)
		if ok {
			for key, val := range configs {
				strval, ok := val.(string)
				if ok && reg.MatchString(strval) {
					arrs := reg.FindStringSubmatch(strval)
					if len(arrs) > 1 {
						this.app.Debug("app", "读取环境变量 %s", arrs[1])
						val = os.Getenv(arrs[1])
					}
				}else{
					//this.app.Debug("app", "正常变量 %s", val)
				}
				moduleConfig.SetItem(key, val)
			}
		}
		this.ModuleConfigs[moduleName] = moduleConfig
	}
	return
}
