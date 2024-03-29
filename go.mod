module github.com/zegl/deconz_exporter

go 1.18

require (
	github.com/jurgen-kluft/go-conbee v0.0.0-20220115123148-2a74bb20e181
	github.com/prometheus/client_golang v1.12.2
	go.uber.org/zap v1.20.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/jurgen-kluft/go-conbee => github.com/zegl/go-conbee v0.0.0-20220718191014-c05f18123550
