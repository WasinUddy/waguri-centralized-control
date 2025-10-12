module waguri-centralized-control/proxy

go 1.25.0

require (
	waguri-centralized-control/packages/go-utils/config v0.0.0
	waguri-centralized-control/packages/go-utils/telemetry v0.0.0
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace waguri-centralized-control/packages/go-utils/config => ../../packages/go-utils/config

replace waguri-centralized-control/packages/go-utils/telemetry => ../../packages/go-utils/telemetry
