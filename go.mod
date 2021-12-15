module github.com/grafana/grafana-starter-datasource-backend

go 1.16

require (
	github.com/alexeymaximov/go-bio v0.1.0
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/grafana/grafana-plugin-sdk-go v0.105.0
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/smallnest/ringbuffer v0.0.0-20210227121335-0a58434b36f2
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2
	golang.org/x/text v0.3.3
)

replace github.com/alexeymaximov/go-bio v0.1.0 => github.com/alexanderzobnin/go-bio v0.1.1-0.20210702075309-d3abeb65f6d4
