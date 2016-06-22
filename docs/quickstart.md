# Quickstart

The following section provides the info regarding how to prepare and use
`ralph-cli`.  After reading it, you'll be able to perform a successful scan of
your iDRAC/iLO host, and save its result to Ralph (at this moment, it will be
limited only to MAC addresses, but that is going to change soon!).

## Getting ralph-cli binary

You can either [download][releases] a pre-built binary for your OS/architecture,
or build it from the source code. In the former case, skip to the
[next sub-section][self-config], otherwise read on.

`ralph-cli` is written in Go and uses [Glide][glide] for managing its
dependencies, so assuming that you already have it installed on your system (see
instructions for this [here][glide-install]), and that you have cloned
`ralph-cli`'s repo to your `$GOPATH`, all you have to do is to issue `glide install`
command and then `go build` (or `go install`, if you prefer to make your Go
binaries that way).

If you are a die-hard Go programmer that uses only stdlib for everything, you can
proceed in a usual way, i.e. by issuing `go get github.com/allegro/ralph-cli`,
but in such case, you need to handle `ralph-cli`'s dependencies by
yourself. This shouldn't be difficult, though.

Either way, after building your binary, you can verify it by issuing
`ralph-cli --help` - it should give you a simple usage screen.

## Configuring ralph-cli

At the initial run, `ralph-cli` creates `~/.ralph-cli` directory, where its
config file (`config.toml`) is stored (along with `scripts` sub-directory - more
on this later). Its contents should look like this:

    Debug = false
    LogOutput = ""
    ClientTimeout = 10
    RalphAPIURL = "change_me"
    RalphAPIKey = "change_me"
    ManagementUserName = "change_me"
    ManagementUserPassword = "change_me"

Before doing anything, you need to replace dummy defaults denoted by
`"change_me` string. For the meaning of those settings, see
[here][concepts-config].

## Running scan scripts

`ralph-cli` [scans][concepts-scan] given IP address by executing one of the
[scripts][concepts-scripts] located in the `~/.ralph-cli/scripts` directory (you
may think of them as "plugins"). By default, it comes with two scripts:
`idrac.py` and `ilo.py`, which are both written in Python. If you have
`python3`, `pip` and `virtualenv` in your `$PATH`, you can continue to the next
paragraph - otherwise consult your package manager's manuals how to install
Python and also check out `pip`'s and `virtualenv`'s docs [here][pip] and
[here][virtualenv].

Having everything ready, we can perform the actual scan. Let's say that you have
some Dell PowerEdge R620 server, with iDRAC exposed on management IP
`11.22.33.44`. Let's suppose that this server is added to Ralph, but in the
"Network" tab it has only this management IP visible - and nothing else. But!
You know that this server has 4 NICs, and you want this information to be
available in Ralph too. Of course, you can log in to said iDRAC and copy relevant
information manually (e.g. MAC addresses of these 4 NICs), but this is rather
cumbersome, error-prone and time-consuming. `ralph-cli` to the rescue! You can
achieve the same thing by issuing:

    ralph-cli scan 11.22.33.44 --script=idrac.py --dry-run

...which would produce output similar to this:

    INFO: Running in dry-run mode, no changes will be saved in Ralph.
    EthernetComponent with MAC address a1:b2:c3:d4:e5:aa created successfully.
    EthernetComponent with MAC address a1:b2:c3:d4:e5:bb created successfully.
    EthernetComponent with MAC address a1:b2:c3:d4:e5:cc created successfully.
    EthernetComponent with MAC address a1:b2:c3:d4:e5:dd created successfully.

Notice that we are running `ralph-cli` in "dry-run" mode, which is a good idea
when you need some sort of control over your data. After examining this output
and finding it OK, you can safely issue the same command without `--dry-run`
switch. After that, you can check that the data was actually sent to Ralph by
going back to aforementioned "Network" tab in Ralph.

You may be wondering what would happen if you'd issue the same command
again. Well, try it and see by yourself! Unless you've changed something in the
hardware of your server (e.g. replaced some network card), you should see this
message:

    No changes detected.

And it means that the state of your server in Ralph reflects its actual state
(at least in terms of components visible to `idrac.py` script).

## Going further

This tutorial gave you the minimum info needed to start using `ralph-cli` for
extracting MAC addresses from a Dell server running iDRAC. But you can go
further than that. There's another script bundled with `ralph-cli` named
`ilo.py`, which is intended for use with HP servers equipped with iLO service
processor. And since both `idrac.py` and `ilo.py` are scripts, not some binary
plugins, you can freely modify them in-place (e.g, to make them suit your needs
better). But the best thing is, that you can actually write your own scripts,
*in any language you want*, as long as they conform to
[Scripts Contract][concepts-contract] and you have means for running them from
your host (e.g. access to interpreter, required libraries etc.).

[self-config]: quickstart.md#configuring-ralph-cli
[concepts-config]: concepts.md#config
[concepts-scan]: concepts.md#scan
[concepts-contract]: concepts.md#scripts-contract
[concepts-scripts]: concepts.md#scripts

[releases]: https://github.com/allegro/ralph-cli/releases
[glide]: https://github.com/Masterminds/glide
[glide-install]: https://github.com/Masterminds/glide#install
[pip]: https://pip.pypa.io/en/stable/installing/
[virtualenv]: https://packaging.python.org/en/latest/installing/#creating-and-using-virtual-environments
