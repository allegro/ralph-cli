# Quickstart

The following section provides the info regarding how to prepare and use
`ralph-cli`.  After reading it, you'll be able to perform a successful scan of
your iDRAC/iLO host, and save its result (i.e. discovered hardware components)
to Ralph.

## Installation

You can download a pre-built, self-contained `ralph-cli` executable from
[here][releases]. If your OS/architecture is not listed there, you can either
request it from us by opening an issue on [our GitHub profile][issues], or build
it from the source code. In the latter case, please refer to
[How to build ralph-cli?][development-build] section.

## Configuring ralph-cli

At the initial run (e.g. `ralph-cli --help`), `ralph-cli` creates `~/.ralph-cli`
directory, where its config file (`config.toml`) is stored (along with `scripts`
sub-directory - more on this later). Its contents should look like this:

```no-highlight
RalphAPIURL = "change_me"
RalphAPIKey = "change_me"
ManagementUserName = "change_me"
ManagementUserPassword = "change_me"
```

Before doing anything real, you need to replace dummy defaults denoted by
`"change_me` string. For the meaning of those settings, which are all required,
see [here][concepts-config].

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
"Network" tab it has only this management IP visible - and nothing else. But you
know that this server has four network cards, and you want this information to
be available in Ralph too. Of course, you can log in to said iDRAC and copy
relevant information manually, but this is rather cumbersome, error-prone and
time-consuming. `ralph-cli` to the rescue! You can achieve the same thing by
issuing:

```no-highlight
ralph-cli scan 11.22.33.44 --script=idrac.py --components=eth --dry-run
```

...which would produce output similar to this:

```no-highlight
INFO: Running in dry-run mode, no changes will be saved in Ralph.
Ethernet{id: 1, base_object_id: 1, mac: a1:b2:c3:d4:e5:aa, model_name: Intel(R) Ethernet 10G 4P X520/I350 rNDC, speed: 10 Gbps, firmware_version: 1.2.3} created successfully.
Ethernet{id: 2, base_object_id: 1, mac: a1:b2:c3:d4:e5:bb, model_name: Intel(R) Ethernet 10G 4P X520/I350 rNDC, speed: 10 Gbps, firmware_version: 1.2.3} created successfully.
Ethernet{id: 3, base_object_id: 1, mac: a1:b2:c3:d4:e5:cc, model_name: Intel(R) Ethernet 10G 4P X520/I350 rNDC, speed: 10 Gbps, firmware_version: 1.2.3} created successfully.
Ethernet{id: 4, base_object_id: 1, mac: a1:b2:c3:d4:e5:dd, model_name: Intel(R) Ethernet 10G 4P X520/I350 rNDC, speed: 10 Gbps, firmware_version: 1.2.3} created successfully.
```

Notice that we are running `ralph-cli` in "dry-run" mode, which is a good idea
when you need some sort of control over your data. After examining this output
and finding it OK, you can safely issue the same command without `--dry-run`
switch. After that, you can check that the data was actually sent to Ralph by
going back to aforementioned "Network" tab in Ralph.

You may be wondering what would happen if you'd issue the same command
again. Well, try it and see by yourself! Unless you've replaced some network
card, you should see this message:

```no-highlight
No changes detected.
```

And it means that the state of your server stored in Ralph reflects its actual
state in regards to network cards visible to `idrac.py` script.

You may want to perform similar detection for other components as well, e.g.:

```no-highlight
ralph-cli scan 11.22.33.44 --script=idrac.py --components=eth,mem,proc
```

...will try to detect network cards, memory and processors. For a possible
arguments for `--components` switch, see the output for `ralph-cli scan --help`
command. By default (i.e. when you don't specify anything with `--components`
switch), `ralph-cli` will look for all components (`--components=all`).

There are two additional switches for `scan` command, which may be useful for
you: `--with-bios-and-firmware` and `--with-model`. Let's see them in action by
issuing this command:


```no-highlight
ralph-cli scan 11.22.33.44 --script=idrac.py --with-bios-and-firmware --with-model
```

...and we should get the output similar to this:


```no-highlight
DataCenterAsset{id: 1, firmware_version: 1.10.10.10, bios_version: 1.2.3} updated successfully.
DataCenterAsset{id: 1, remarks: >>> ralph-cli: detected model name: Dell PowerEdge R620 <<<} updated successfully.
```

In the above example, the former switch detects firmware and BIOS versions of
your iDRAC server, and writes them to Ralph - they can be examined by looking at
"Basic info" tab in Ralph's GUI.

The latter tries to detect the model name of your server, but appends it into
the "Remarks" field (see "Basic info" tab once again) into a specially formatted
string denoted by `>>> <<<` marks. The rationale behind this is that we don't
want to pollute the actual "Model" field in "Basic info" tab, with values that
may be slightly different due to some arbitrary changes in formatting
(e.g. "Dell PowerEdge R620" and "PowerEdge R620" are essentially the same
models, but different strings), so we think it's better to leave it as a kind of
suggestion for the person, who maintains this data in Ralph.


## Going further

This tutorial gave you the minimum info needed to start using `ralph-cli` for
detecting components of a Dell server running iDRAC. But you can go further than
that. There's another script bundled with `ralph-cli` named `ilo.py`, which is
intended for use with HP servers equipped with iLO service processor (and gives
slightly less info than iDRAC). And since both `idrac.py` and `ilo.py` are
scripts, not some binary plugins, you can freely modify them in-place (e.g, to
make them suit your needs better). But the best thing is, that you can actually
write your own scripts, *in any language you want*, as long as they conform to
[Scripts Contract][concepts-contract] and you have means for running them from
your host (e.g. access to interpreter, required libraries etc.).

[self-config]: quickstart.md#configuring-ralph-cli
[concepts-config]: concepts.md#config
[concepts-scan]: concepts.md#scan
[concepts-contract]: concepts.md#scripts-contract
[concepts-scripts]: concepts.md#scripts
[development-build]: development.md#how-to-build-ralph-cli

[releases]: https://github.com/allegro/ralph-cli/releases
[issues]: https://github.com/allegro/ralph-cli/issues
[pip]: https://pip.pypa.io/en/stable/installing/
[virtualenv]: https://packaging.python.org/en/latest/installing/#creating-and-using-virtual-environments
