# bundled_scripts

This directory contains files (scripts and manifests) for `ralph-cli scan`
command. They are bundled with binary produced by `go build`, by utilising
mechanisms provided by `go-bindata` ([GitHub][1], [GoDoc][2]), so every
change to them should be followed by
`go-bindata -ignore README\\.md -prefix "bundled_scripts" bundled_scripts/`
(issued from the parent dir) - this will generate a `bindata.go` file,
which shouldn't be edited manually.

If you don't have `go-bindata` executable in your `$GOPATH`, then
install it with `go get -u github.com/jteeuwen/go-bindata/...` (it is
not included in `glide.yaml` because it is not dependency of
`ralph-cli` per se).

[1]: https://github.com/jteeuwen/go-bindata
[2]: https://godoc.org/github.com/jteeuwen/go-bindata
