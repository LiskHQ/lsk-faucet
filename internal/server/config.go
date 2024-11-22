package server

type Config struct {
	network         string
	symbol          string
	payout          float64
	tokenDecimals   int
	httpPort        int
	interval        int
	proxyCount      int
	hcaptchaSiteKey string
	hcaptchaSecret  string
	explorerURL     string
	explorerTxPath  string
}

func NewConfig(network, symbol string, payout float64, tokenDecimals, httpPort, interval, proxyCount int, hcaptchaSiteKey, hcaptchaSecret, explorerURL, explorerTxPath string) *Config {
	return &Config{
		network:         network,
		symbol:          symbol,
		httpPort:        httpPort,
		interval:        interval,
		payout:          payout,
		tokenDecimals:   tokenDecimals,
		proxyCount:      proxyCount,
		hcaptchaSiteKey: hcaptchaSiteKey,
		hcaptchaSecret:  hcaptchaSecret,
		explorerURL:     explorerURL,
		explorerTxPath:  explorerTxPath,
	}
}
