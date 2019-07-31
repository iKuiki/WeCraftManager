module wecraftmanager

go 1.12

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/google/uuid v1.1.1
	github.com/ikuiki/go-component v0.0.0-20171218165758-b9f2562e71d1
	github.com/ikuiki/storer v1.0.0
	github.com/ikuiki/wwdk v2.5.0+incompatible
	github.com/liangdas/mqant v1.8.1
	github.com/pkg/errors v0.8.1
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	wegate v0.0.0-00010101000000-000000000000
)

replace wegate => github.com/ikuiki/wegate v1.0.4-0.20190731072427-b0a1a1b7b7f2

replace github.com/liangdas/mqant => github.com/ikuiki/mqant v1.8.1-0.20190427142930-7dabfa32d064
