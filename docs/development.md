# Development

This section is meant for people who are willing to contribute to `ralph-cli` or
just want to experiment with it (or just build `ralph-cli` from source). Here
you'll also find our plans for the future.

## How to build ralph-cli?

`ralph-cli` is written in Go and uses [Glide][glide] for managing its
dependencies, so assuming that you already have Glide installed on your system
(see instructions for this [here][glide-install]), and that you have cloned
`ralph-cli`'s repo to your `$GOPATH`, all you have to do is to issue `glide
install` command and then `go build` (or `go install`, if you prefer to make
your Go binaries that way).

If you are a die-hard Go programmer that uses only stdlib for everything, you can
proceed in a usual way, i.e. by issuing `go get github.com/allegro/ralph-cli`,
but in such case, you need to handle `ralph-cli`'s dependencies by
yourself. This shouldn't be difficult, though.

Either way, after building your binary, you can verify it by issuing
`ralph-cli --help` - it should give you a simple usage screen.

## Bundled scripts

`ralph-cli` comes with two default scripts for the `scan` command - `idrac.py`
and `ilo.py`. They are written in Python, and embedded into `ralph-cli`'s binary
by `go-bindata` tool. If you'd find yourself in need of making any permanent
change to them, please refer to `bundled_scripts/README.md` file for details on
how to add them to resulting `ralph-cli` binary.

## Ideas for Future Development

Here are some of the ideas that we are working on, or that may be implemented in
`ralph-cli`. It should give you an approximate picture in which direction
`ralph-cli`'s development is heading.

### Sooner

* Ability to refresh/recreate virtualenvs used by scan scripts written in Python
  (e.g. after adding new dependency to manifest file).
* Integration with [Logstash][logstash] (this is almost ready, though).
* Ability to feed `ralph-cli scan` with ready-made JSON files (i.e. without
  launching any scan scripts).
* Ability to update all components detected by scan on a given host at once
  (i.e. with a single HTTP request over a single API endpoint).
* Ability to configure `ralph-cli` by environment variables, which would take
  precedence over the config file.
* Ability to scan whole range of hosts/networks (at this moment, `ralph-cli`
  operates only on single hosts).
* Support for Windows.
* Some minor improvements like setting timeouts for scan, adding progress bars etc.

### Later

* Ability to add/checkout scan scripts to/from a Git repo without direct
  interaction with Git itself, e.g. `ralph-cli script --commit`,
  `ralph-cli script --checkout` - this should include commands like
  `ralph-cli script --edit` and so on.
* Ability to deploy hosts without touching Ralph's GUI, e.g. `ralph-cli deploy`.

[glide]: https://github.com/Masterminds/glide
[glide-install]: https://github.com/Masterminds/glide#install
[logstash]: https://www.elastic.co/products/logstash
