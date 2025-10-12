module waguri-centralized-control/dns

go 1.25.0

require (
	waguri-centralized-control/packages/go-utils/config v0.0.0
	waguri-centralized-control/packages/go-utils/telemetry v0.0.0
)

require (
	github.com/miekg/dns v1.1.68 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace waguri-centralized-control/packages/go-utils/config => ../../packages/go-utils/config

replace waguri-centralized-control/packages/go-utils/telemetry => ../../packages/go-utils/telemetry
