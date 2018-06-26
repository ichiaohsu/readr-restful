package config

import (
	"log"

	"github.com/spf13/viper"
)

var Config appConfig

type appConfig struct {
	SQL struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	} `mapstructure:"sql"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		Channels struct {
			TalkComment string `mapstructure:"talk_comments"`
		} `mapstructure:"channels"`
	} `mapstructure:"redis"`

	Crawler struct {
		Headers map[string]string `mapstructure:"headers"`
	} `mapstructure:"crawler"`

	Mail struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		UserName string `mapstructure:"user_name"`
	} `mapstructure:"mail"`

	SearchFeed struct {
		AppID     string `mapstructure:"app_id"`
		AppKey    string `mapstructure:"app_key"`
		IndexName string `mapstructure:"index_name"`
		MaxRetry  int    `mapstructure:"max_retry"`
	} `mapstructure:"search_feed"`

	Straats struct {
		AppID     string `mapstructure:"app_id"`
		AppKey    string `mapstructure:"app_key"`
		APIServer string `mapstructure:"api_server"`
	} `mapstructure:"straats"`

	Models struct {
		Members               map[string]int `mapstructure:"members"`
		Posts                 map[string]int `mapstructure:"posts"`
		PostType              map[string]int `mapstructure:"post_type"`
		PostPublishStatus     map[string]int `mapstructure:"post_publish_status"`
		Tags                  map[string]int `mapstructure:"tags"`
		ProjectsActive        map[string]int `mapstructure:"projects_active"`
		ProjectsStatus        map[string]int `mapstructure:"projects_status"`
		ProjectsPublishStatus map[string]int `mapstructure:"projects_publish_status"`
		Memos                 map[string]int `mapstructure:"memos"`
		MemosPublishStatus    map[string]int `mapstructure:"memos_publish_status"`
		Comment               map[string]int `mapstructure:"comment"`
		CommentStatus         map[string]int `mapstructure:"comment_status"`
		ReportedCommentStatus map[string]int `mapstructure:"reported_comment_status"`
		Reports               map[string]int `mapstructure:"reports"`
		ReportsPublishStatus  map[string]int `mapstructure:"reports_publish_status"`
		FollowingType         map[string]int `mapstructure:"following_type"`
	} `mapstructure:"models"`
}

func LoadConfig(configPath string, configName string) error {

	v := viper.New()
	v.SetConfigType("json")

	if configPath != "" {
		v.AddConfigPath(configPath)
	} else {
		// Default path
		v.AddConfigPath("./config")
	}

	if configName != "" {
		v.SetConfigName(configName)
	} else {
		v.SetConfigName("main")
	}

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
		return err
	}
	log.Println("Using config file:", v.ConfigFileUsed())

	if err := v.Unmarshal(&Config); err != nil {
		log.Fatalf("Error unmarshal config file, %s", err)
		return err
	}
	return nil
}