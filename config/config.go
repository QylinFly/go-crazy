/**
 * 配置
 * File: config.go
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 2, 2017-12-19 8:39:49 pm
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 2, 2017-12-19 8:40:14 pm
 * -----
 * Copyright 2017 - 2027 乐编程, 乐编程
 */


 package Config


var Config = struct {
	APPName string `default:"app name"`
	Port string `default:"80"`
	Type string `default:"provider"`
	Channels int `default:"200"`

	DubboPort int `default:"20880"`

	EtcdUrl string `default:"172.17.0.1:2379"`

	DB_CONNECTION string `default:"mysql"`
	DB struct {
		Host 	 string	`default:"127.0.0.1"`
		Name     string
		User     string `default:"root"`
		Password string `required:"true" env:"DBPassword"`
		Port     string   `default:"3306"`
	}

	Contacts []struct {
		Name  string
		Email string `required:"true"`
	}
}{}
