package configs


type Config struct {
	LogAcquisitionAddr string `yaml:"log_acquisition_addr"` // 日志获取地址
	LogDeliveryAddr    string `yaml:"log_delivery_addr"`    //日志投递地址
}
