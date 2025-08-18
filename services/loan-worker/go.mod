module github.com/huuhoait/los-demo/services/loan-worker

go 1.21

require (
	github.com/conductor-sdk/conductor-go v1.5.4
	github.com/google/uuid v1.4.0
	github.com/huuhoait/los-demo/services/shared v0.0.0
	github.com/lib/pq v1.10.9
	go.uber.org/zap v1.27.0
)

replace github.com/huuhoait/los-demo/services/shared => ../shared

require (
	github.com/antihax/optional v1.0.0 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.2.2 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
