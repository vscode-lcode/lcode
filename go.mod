module github.com/vscode-lcode/lcode/v2

go 1.19

require (
	github.com/SierraSoftworks/multicast/v2 v2.0.0
	github.com/alessio/shellescape v1.4.1
	github.com/google/uuid v1.3.0
	github.com/jellydator/ttlcache/v3 v3.0.1
	github.com/lainio/err2 v0.8.13
	github.com/mattn/go-sqlite3 v1.14.16
	go.opentelemetry.io/otel v1.11.2
	go.opentelemetry.io/otel/sdk v1.11.2
	go.opentelemetry.io/otel/trace v1.11.2
	golang.org/x/net v0.5.0
	xorm.io/builder v0.3.12
	xorm.io/xorm v1.3.2
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.8.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.4.0 // indirect
)

// replace golang.org/x/net/webdav => ../net-webdav
