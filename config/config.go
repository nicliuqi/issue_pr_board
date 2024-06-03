package config

import (
	"os"
	"sigs.k8s.io/yaml"
)

var AppConfig = &appConfig{}

type appConfig struct {
	AccessToken      string `json:"access_token"`
	DBChar           string `json:"db_char"`
	DBHost           string `json:"db_host"`
	DBName           string `json:"db_name"`
	DBPassword       string `json:"db_password"`
	DBPort           int    `json:"db_port"`
	DBUsername       string `json:"db_username"`
	EnterpriseId     string `json:"enterprise_id"`
	GiteeV5ApiPrefix string `json:"gitee_api_v5_prefix"`
	GiteeV8ApiPrefix string `json:"gitee_api_v8_prefix"`
	RandRawString    string `json:"rand_raw_string"`
	SMTPHost         string `json:"smtp_host"`
	SMTPPassword     string `json:"smtp_password"`
	SMTPPort         int    `json:"smtp_port"`
	SMTPSender       string `json:"smtp_sender"`
	SMTPUsername     string `json:"smtp_username"`
	V8Token          string `json:"v8_token"`
	VerifyInterval   int    `json:"verify_interval"`
	VerifyExpire     int    `json:"verify_expire"`
	WebhookToken     string `json:"webhook_token"`
}

func InitAppConfig(path string) error {
	cfg := AppConfig
	if err := LoadFromYaml(path, cfg); err != nil {
		return err
	}
	cfg.setDefault()
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromYaml(path string, cfg interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := []byte(os.ExpandEnv(string(b)))

	if err = yaml.Unmarshal(content, cfg); err != nil {
		return err
	}
	return err
}

func (cfg *appConfig) setDefault() {}
