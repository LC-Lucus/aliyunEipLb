package config

import (
	"aliyunEipLb/aliyun"
	"aliyunEipLb/log"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

// 阿里云配置项
var AliyunConfig = new(aliyun.ALiYun)
var AliyunConfigs []*aliyun.ALiYun

// 数据库配置项
var cfgDatabase *viper.Viper

func Setup(path string) {
	viper.SetConfigFile(path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(fmt.Sprintf("Read config file fail: %s", err.Error()))
	}
	err = viper.ReadConfig(strings.NewReader(os.ExpandEnv(string(content))))
	if err != nil {
		log.Fatal(fmt.Sprintf("Parse config file fail: %s", err.Error()))
	}

	// 数据库配置
	cfgDatabase = viper.Sub("settings.database")
	if cfgDatabase == nil {
		panic("No found settings.database in the configuration")
	}
	DatabaseConfig = InitDatabase(cfgDatabase)
	DB = ConnectMysql(DatabaseConfig.Driver, DatabaseConfig.Source)

	// 阿里云配置
	cfgAliyun := viper.Get("settings.aliyun")
	for _, r := range cfgAliyun.([]interface{}) {
		ali := r.(map[string]interface{})
		var region []string
		for _, i := range ali["region"].([]interface{}) {
			region = append(region, i.(string))
		}
		AliyunConfig = InitAliyun(
			ali["ak"].(string),
			ali["sk"].(string),
			ali["account"].(string),
			region,
		)
		AliyunConfigs = append(AliyunConfigs, AliyunConfig)
	}

	//if cfgAliyun == nil {
	//	panic("No found settings.aliyun in the configuration")
	//}
	//AliyunConfig = InitAliyun(cfgAliyun)
}
func InitAliyun(ak, sk, account string, region []string) *aliyun.ALiYun {
	aliyunClient := &aliyun.ALiYun{
		AK:      ak,
		SK:      sk,
		Account: account,
		Region:  region,
	}
	return aliyunClient
}

//func InitAliyun(cfg *viper.Viper) *aliyun.ALiYun {
//	aliyunClient := &aliyun.ALiYun{
//		AK:     cfg.GetString("AK"),
//		SK:     cfg.GetString("SK"),
//		Region: cfg.GetStringSlice("Region"),
//	}
//	return aliyunClient
//}
