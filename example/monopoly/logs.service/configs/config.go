package configs

import "github.com/yamakiller/velcro-go/cluster/elastic"

type Config struct {
	Elastic elastic.ElasticConfig `yaml:"elastic"`
}
