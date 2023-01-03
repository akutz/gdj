module github.com/akutz/gdj/canaries

go 1.17

replace github.com/akutz/gdj => ../

require (
	github.com/akutz/gdj v0.0.0-20230102150933-91e2b897db3e
	github.com/google/go-cmp v0.5.9
	github.com/stretchr/testify v1.8.1
	github.com/vmware/govmomi v0.28.1-0.20221215183112-aca02acc48b0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
