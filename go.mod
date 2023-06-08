module github.com/streamdal/shopify-sarama

go 1.17

require (
	github.com/Shopify/sarama v1.38.1
	github.com/Shopify/toxiproxy/v2 v2.5.0
	github.com/davecgh/go-spew v1.1.1
	github.com/eapache/go-resiliency v1.3.0
	github.com/eapache/go-xerial-snappy v0.0.0-20230111030713-bf00bc1b83b6
	github.com/eapache/queue v1.1.0
	github.com/fortytw2/leaktest v1.3.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jcmturner/gofork v1.7.6
	github.com/jcmturner/gokrb5/v8 v8.4.3
	github.com/klauspost/compress v1.15.14
	github.com/pierrec/lz4/v4 v4.1.17
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/stretchr/testify v1.8.1
	github.com/xdg-go/scram v1.1.2
	golang.org/x/net v0.7.0
	golang.org/x/sync v0.1.0
)

require (
	github.com/CosmWasm/tinyjson v0.9.0 // indirect
	github.com/batchcorp/plumber-schemas v0.0.183-0.20230607212720-e58a068de32c // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magefile/mage v1.14.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/streamdal/pii v0.0.4-0.20230607185452-2896fc3a28a4 // indirect
	github.com/tetratelabs/wazero v1.2.0 // indirect
	github.com/tidwall/gjson v1.14.1 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/wasilibs/go-re2 v1.1.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20200825200019-8632dd797987 // indirect
	google.golang.org/grpc v1.40.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/streamdal/dataqual v0.0.18

retract (
	v1.32.0 // producer hangs on retry https://github.com/Shopify/sarama/issues/2150
	[v1.31.0, v1.31.1] // producer deadlock https://github.com/Shopify/sarama/issues/2129
	[v1.26.0, v1.26.1] // consumer fetch session allocation https://github.com/Shopify/sarama/pull/1644
	[v1.24.1, v1.25.0] // consumer group metadata reqs https://github.com/Shopify/sarama/issues/1544
)
