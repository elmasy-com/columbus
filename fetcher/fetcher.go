package fetcher

var (

	// From https://www.gstatic.com/ct/log_list/log_list.json
	Logs = []log{
		{uri: "https://ct.googleapis.com/logs/argon2022/", name: "argon2022"},
		{uri: "https://ct.googleapis.com/logs/argon2023/", name: "argon2023"},
		{uri: "https://ct.googleapis.com/logs/xenon2022/", name: "xenon2022"},
		{uri: "https://ct.googleapis.com/logs/xenon2023/", name: "xenon2023"},
		{uri: "https://ct.googleapis.com/icarus/", name: "icarus"},
		{uri: "https://ct.googleapis.com/pilot/", name: "pilot"},
		{uri: "https://ct.googleapis.com/rocketeer/", name: "rocketeer"},
		{uri: "https://ct.googleapis.com/skydiver/", name: "skydiver"},
		{uri: "https://ct.cloudflare.com/logs/nimbus2022/", name: "nimbus2022"},
		{uri: "https://ct.cloudflare.com/logs/nimbus2023/", name: "nimbus2023"},
		//{uri: "https://ct1.digicert-ct.com/log/", name: "digicert-ct1"},
		//{uri: "https://ct2.digicert-ct.com/log/", name: "digicert-ct2"},
		{uri: "https://yeti2022.ct.digicert.com/log/", name: "yeti2022"},
		{uri: "https://yeti2022-2.ct.digicert.com/log/", name: "yeti2022-2"},
		{uri: "https://yeti2023.ct.digicert.com/log/", name: "yeti2023"},
		{uri: "https://nessie2022.ct.digicert.com/log/", name: "nessie2022"},
		{uri: "https://nessie2023.ct.digicert.com/log/", name: "nessie2023"},
		{uri: "https://sabre.ct.comodo.com/", name: "sabre"},
		{uri: "https://mammoth.ct.comodo.com/", name: "mammoth"},
		//{uri: "https://oak.ct.letsencrypt.org/2019/", name: "oak2019"},
		//{uri: "https://oak.ct.letsencrypt.org/2020/", name: "oak2020"},
		//{uri: "https://oak.ct.letsencrypt.org/2021/", name: "oak2021"},
		{uri: "https://oak.ct.letsencrypt.org/2022/", name: "oak2022"},
		{uri: "https://oak.ct.letsencrypt.org/2023/", name: "oak2023"},
		{uri: "https://ct.trustasia.com/log2022/", name: "trustasia2022"},
		{uri: "https://ct.trustasia.com/log2023/", name: "trustasia2023"},
	}
)

// Start fetching every log.
func Start() {

	for i := range Logs {
		Logs[i].Start()
	}
}

// Close every log. The closing function of logs blosk until it is closed.
func Close() {

	for i := range Logs {
		Logs[i].Close()
	}
}
