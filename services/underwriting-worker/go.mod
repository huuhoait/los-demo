module underwriting_worker

go 1.23.3

require (
	github.com/huuhoait/los-demo/services/shared v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/huuhoait/los-demo/services/shared => ../shared

require (
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)
