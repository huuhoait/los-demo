module github.com/huuhoait/los-demo/services/loan-worker

go 1.21

require (
	github.com/conductor-sdk/conductor-go v1.5.4
	github.com/google/uuid v1.3.0
	github.com/huuhoait/los-demo/services/shared v0.0.0
	github.com/lib/pq v1.10.9
	github.com/nicksnyder/go-i18n/v2 v2.2.2
	github.com/pelletier/go-toml/v2 v2.0.8
	go.uber.org/zap v1.25.0
	golang.org/x/text v0.21.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/huuhoait/los-demo/services/shared => ../shared

require (
	github.com/antihax/optional v1.0.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
)
