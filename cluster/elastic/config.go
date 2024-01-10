package elastic

type ElasticConfig struct {
	ElasticAddress string   `yaml:"elastic_address"`// elastic 连接地址
	ElasticKafka KafkaConfig `yaml:"elastic_kafka"`
}

type KafkaConfig struct {
	Broker    string   `yaml:"broker"`  // kakfa 连接地址
	Topics     []string `yaml:"topics"`   // kakfa 主题
	Group     string   `yaml:"group"`   // kakfa  组
	Version   string   `yaml:"version"` // kakfa 版本
}

