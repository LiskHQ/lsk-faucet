package server

type Config struct {
	network         string
	symbol          string
	httpPort        int
	interval        int
	payout          int
	proxyCount      int
	hcaptchaSiteKey string
	hcaptchaSecret  string
	explorerURL     string
	explorerTxPath  string
}

func NewConfig(network, symbol string, httpPort, interval, payout, proxyCount int, hcaptchaSiteKey, hcaptchaSecret, explorerURL, explorerTxPath string) *Config {
	return &Config{
		network:         network,
		symbol:          symbol,
		httpPort:        httpPort,
		interval:        interval,
		payout:          payout,
		proxyCount:      proxyCount,
		hcaptchaSiteKey: hcaptchaSiteKey,
		hcaptchaSecret:  hcaptchaSecret,
		explorerURL:     explorerURL,
		explorerTxPath:  explorerTxPath,
	}
}
