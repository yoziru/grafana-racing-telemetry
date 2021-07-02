module github.com/grafana/grafana-starter-datasource-backend

go 1.16

require (
	github.com/alexeymaximov/go-bio v0.1.0
	github.com/grafana/grafana-plugin-sdk-go v0.105.0
	github.com/magefile/mage v1.11.0 // indirect
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2
)

replace github.com/alexeymaximov/go-bio v0.1.0 => github.com/alexanderzobnin/go-bio v0.1.1-0.20210702075309-d3abeb65f6d4
