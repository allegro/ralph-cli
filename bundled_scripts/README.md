# bundled_scripts

This directory contains files (scripts and manifests) for `ralph-cli
scan` command. They are bundled with binary produced by `go build`, by
utilising mechanisms provided by `github.com/jteeuwen/go-bindata`, so
every change to them should be followed by `go-bindata -ignore
README\\.md -prefix "bundled_scripts" bundled_scripts/` (issued from
the parent dir) - this will generate a `bindata.go` file, which
shouldn't be edited manually.

If you don't have `go-bindata` executable in your `$GOPATH`, then
install it with `go get -u github.com/jteeuwen/go-bindata/...` (it is
not included in `glide.yaml` because it is not dependency of
`ralph-cli` per se).
