package config

import "github.com/opensourceways/app-cla-server/util"

var AppConfig = &appConfig{}

type appConfig struct {
	AccessToken	string	`json:"access_token"`
	DBChar		string	`json:"db_char"`
	DBHost		string	`json:"db_host"`
	DBName		string	`json:"db_name"`
	DBPassword	string	`json:"db_password"`
	DBPort		string	`json:"db_port"`
	DBUsername	string	`json:"db_username"`
	EnterpriseId	string	`json:"enterprise_id"`
	SMTPHost	string	`json:"smtp_host"`
	SMTPPort	string	`json:"smtp_port"`
	SMTPUsername	string	`json:"smtp_username"`
	SMTPPassword	string	`json:"smtp_password"`
	V8Token		string	`json:"v8_token"`
	VerifyInterval	int	`json:"verify_interval"`
	VerifyExpire	int	`json:"verify_expire"`
}

func InitAppConfig(path string) error {
	cfg := AppConfig
	if err := uitl.LoadFromYaml(path, cfg); err != nil {
		return err
	}
	cfg.setDefault()
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func (cfg *appConfig) setDefault() {}
