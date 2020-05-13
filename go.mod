module github.com/figment-networks/polkadothub-indexer

go 1.13

replace github.com/figment-networks/polkadothub-proxy/grpc v0.0.0 => ../polkadothub-proxy/grpc

require (
	github.com/aristanetworks/goarista v0.0.0-20200410125653-0a3087568c00 // indirect
	github.com/centrifuge/go-substrate-rpc-client v1.1.1-0.20200408213042-6f0547f2818d
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/ethereum/go-ethereum v1.9.13 // indirect
	github.com/figment-networks/polkadothub-proxy/grpc v0.0.0
	github.com/gin-gonic/gin v1.5.0
	github.com/golang/protobuf v1.4.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/jinzhu/gorm v1.9.12
	github.com/lib/pq v1.3.0
	github.com/pierrec/xxHash v0.1.5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/rollbar/rollbar-go v1.2.0
	github.com/rs/cors v1.7.0 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	go.uber.org/zap v1.14.0
	golang.org/x/crypto v0.0.0-20200420201142-3c4aac89819a // indirect
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.22.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
)
