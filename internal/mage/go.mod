module github.com/dagger/dagger/internal/mage

go 1.20

require (
	dagger.io/dagger v0.10.3
	github.com/hexops/gotextdiff v1.0.3
	github.com/magefile/mage v1.15.0
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
	golang.org/x/mod v0.14.0
	golang.org/x/sync v0.6.0
)

require (
	github.com/99designs/gqlgen v0.17.33 // indirect
	github.com/Khan/genqlient v0.6.0 // indirect
	github.com/adrg/xdg v0.4.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/vektah/gqlparser/v2 v2.5.6 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

// needed to resolve "ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules"
replace cloud.google.com/go => cloud.google.com/go v0.100.2

replace github.com/dagger/dagger => ../../
