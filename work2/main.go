package main

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	// initViper()
	initViperRemote()
	// // initViperWatch()
	server := InitWebServer()
	server.Run(":8081")
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func initViper() {

	cfile := pflag.String("config",
		"config/config.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)

	// viper.SetConfigName("dev")
	// // 当前工作目录的config子目录
	// viper.AddConfigPath("config")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	log.Println(viper.Get("test.key"))
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3",
		"localhost:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println("远程配置中心数据发生变更")
	})
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			err = viper.WatchRemoteConfig()
			if err != nil {
				panic(err)
			}
			log.Println("watch", viper.GetString("test.key"))
			time.Sleep(time.Second * 3)
		}
	}()

}

func initViperWatch() {
	cfile := pflag.String("config",
		"config/config.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println(viper.GetString("test.key"))
	})

	// viper.SetConfigName("dev")
	// // 当前工作目录的config子目录
	// viper.AddConfigPath("config")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	log.Println(viper.Get("test.key"))
}
