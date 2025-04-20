module github.com/HUAHUAI23/simple-waf/coraza-spoa

go 1.24.1

require (
	github.com/HUAHUAI23/simple-waf/pkg v0.0.0-20250308163638-ae40316258d8
	github.com/corazawaf/coraza-coreruleset v0.0.0-20240226094324-415b1017abdc
	github.com/corazawaf/coraza/v3 v3.3.2
	github.com/dropmorepackets/haproxy-go v0.0.5
	github.com/jcchavezs/mergefs v0.1.0
	github.com/rs/zerolog v1.33.0
	go.mongodb.org/mongo-driver/v2 v2.1.0
	gopkg.in/yaml.v3 v3.0.1
	istio.io/istio v0.0.0-20240218163812-d80ef7b19049
)

require (
	github.com/alibaba/sentinel-golang v1.0.4 // indirect
	github.com/corazawaf/libinjection-go v0.2.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/magefile/mage v1.15.1-0.20241126214340-bdc92f694516 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/petar-dambovaliev/aho-corasick v0.0.0-20240411101913-e07a1f0e8eb4 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/valllabh/ocsf-schema-golang v1.0.3 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	rsc.io/binaryregexp v0.2.0 // indirect
)

replace github.com/HUAHUAI23/simple-waf/pkg => ../pkg
