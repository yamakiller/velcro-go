package configs

type Config struct{
	Server 				 Server `yaml:"server"`
	Redis                Redis `yaml:"redis"`                     //redis配置
}