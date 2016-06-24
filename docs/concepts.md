# Key Concepts

In order to make your interaction with `ralph-cli` easier, we need to establish
some common vocabulary - hence the following section, where key concepts behind
`ralph-cli` are explained.

## Config

`ralph-cli` has its own configuration file `~/.ralph-cli/config.toml`, in
[TOML][] format. This file is created automatically at the first run of
`ralph-cli`, and it contains some default settings, as presented below:

```no-highlight
Debug = false
LogOutput = ""
ClientTimeout = 10
RalphAPIURL = "change_me"
RalphAPIKey = "change_me"
ManagementUserName = "change_me"
ManagementUserPassword = "change_me"
```

The first three of them (`Debug`, `LogOutput` and `ClientTimeout`) are not used
at this moment, so for now, you can safely ignore their existence (although we
are going to turn them on quite soon). Quite important are the ones with dummy
values denoted by `"change_me"` string - these are required for `ralph-cli` to
operate, and their meaning is as follows:

* `RalphAPIURL` - this should be a string with an URL to your Ralph instance,
  with `/api` path appended, e.g. `"https://my-ralph-instance.local/api"`
* `RalphAPIKey` - your personal API key that `ralph-cli` should use for
  accessing Ralph; you can find this key by visiting Ralph's GUI (go to the
  top-right corner, click on your user name and then select "My profile" - in
  "Personal info" section you should see a field named "API Token" - that's it).
* `ManagementUserName` - user name that is used to access iDRAC/iLO, which
  `ralph-cli` exposes to scan scripts as `MANAGEMENT_USER_NAME` environment
  variable (see [Scripts Contract][self-contract])
* `ManagementUserPassword` - password for the above, exposed as
  `MANAGEMENT_USER_PASSWORD` environment variable

Please note that due to the presence of credentials in `config.toml`, this file
should remain readable only to its owner (it is `0600` by default, so you don't
have to do anything) - otherwise `ralph-cli` will refuse to cooperate.

Also, keep in mind that this is just an initial structure of `ralph-cli` config
file, and it is subject to change until version `1.0.0` is reached.

## Scan

Scan is one of the commands available via `ralph-cli` (well, at this moment,
this is the only command available, but we are going to add more - see section
[Ideas for Future Development][ideas]). The idea behind this is simple:

1. Access a host given as an IP address (via iDRAC, iLO, Puppet, SSH or whatever
   method you'd find useful).
2. Gather some info regarding its configuration (hardware components, software
   etc. - see [Scripts Contract][self-contract]).
3. Process it in some way (e.g. find a difference between what has been
   discovered on this host and what is stored in Ralph, and mark it for
   update/deletion).
4. Send it to Ralph.

The first two steps are handled by scan scripts (see next section), the last two
are handled exclusively by `ralph-cli`, freeing you from the extra work
associated with communication with Ralph.

## Scan scripts

Scripts are the meat of the `scan` command (see previous section). You may think
of them as "plugins", although they are called "scripts" intentionally, to
encourage modifications. Scan scripts can be written in *any* language you want,
as long as they conform to [Scripts Contract][self-contract], and you have means
for running them from your host (e.g. access to interpreter, required libraries
etc.). They should be placed in `~/.ralph-cli/scripts` directory, with an
optional manifest file (see next section).

If you write your scripts in Python, you may need so-called
["virtual environments"][virtualenv] for them - but you don't have to create
and/or activate them by hand, `ralph-cli` will do that for you, as long as you
provide a manifest file listing packages that your script requires. If your Python
script doesn't require anything, then fine, no manifest (and hence no virtualenv)
is needed.

`ralph-cli` comes with two default scripts - `idrac.py` and `ilo.py` - which
should give you an idea what scan scripts should do. You are encouraged to
experiment with them (e.g. by modifying them in-place) - don't worry if you
break them or something - when you delete one of those default scripts (or their
manifests or virtualenvs), `ralph-cli` will restore them to the default
state. Be aware, though, that this behavior doesn't apply to your own scripts
(i.e., with different names that `idrac.py` and `ilo.py`).

## Manifests

These are the files in [TOML][] format that contain some meta data needed for
launching scan scripts. At this moment, they are used exclusively for Python
scripts, and their structure will probably change a lot in the future. Here's an
example of such file, `idrac.toml`:

```no-highlight
Language = "python"
LanguageVersion = 3

[[requirement]]
name = "requests"
version = "2.10.0"
```

Its name (`idrac.toml`) states that this manifest is for `idrac.py` script, and
the fields visible above are rather self-explanatory - `idrac.py` requires
Python 3 and `requests` library in version `2.10.0` (if this field is omitted,
then the newest available version will be used). If you need to specify more
requirements, just add them as another `[[requirement]]` entry, e.g.:

```no-highlight
Language = "python"
LanguageVersion = 3

[[requirement]]
name = "requests"
version = "2.10.0"

[[requirement]]
name = "pyaml"
version = "15.8.2"
```

Manifest files are optional (yet), assuming that your Python script doesn't need
any extra packages, apart from the ones provided by the standard library.

## Scripts Contract

As mentioned in the [Going further][quickstart-further] section, scan scripts
can be written in any language you want, as long as you have means for launching
them. But in order to tame this diversity, they need some well-defined way to
communicate with `ralph-cli` binary. And here comes our Scripts Contract, which
describes exactly this.

### Input

Each script should receive its input parameters only from environment
variables. At this moment, there are three of them:

* `IP_TO_SCAN` - an IP address of a host that we want to scan
* `MANAGEMENT_USER_NAME` - user name that is used to access iDRAC/iLO
* `MANAGEMENT_USER_PASSWORD` - password for the above

`ralph-cli` executes scan scripts as sub-processes, with their own set of
environment variables, so they won't be visible in the parent shell (i.e., the
one from which `ralph-cli` is started).

### Output

Each script should print discovered data on its `stdout`, in the form of a
stringified JSON, as in example presented below:

```no-highlight
{
    "model_name": "Dell PowerEdge R620",
    "processors": [
        {
            "model_name": "Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz",
            "family": "B3",
            "label": "Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz",
            "index": 1,
            "speed": 3600,
            "cores": 8
        },
        {
            "model_name": "Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz",
            "family": "B3",
            "label": "Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz",
            "index": 2,
            "speed": 3600,
            "cores": 8
        }
    ],
    "mac_addresses": [
        "AA:AA:AA:AA:AA:AA",
        "AA:BB:CC:DD:EE:FF",
        "A1:B2:C3:D4:E5:F6",
        "74:86:7A:EE:20:E8"
    ],
    "disks": [
        {
            "model_name": "ATA Samsung SSD 840",
            "family": "ATA",
            "label": "ATA Samsung SSD 840",
            "size": 476,
            "serial_number": "S1AXNSAD8000000"
        },
        {
            "model_name": "ATA Samsung SSD 840",
            "family": "ATA",
            "label": "ATA Samsung SSD 840",
            "size": 476,
            "serial_number": "S1AXNSAD8000001"
        }
    ],
    "serial_number": "UUUZZZ1",
    "memory": [
        {
            "label": "Samsung DDR3 DIMM",
            "speed": 1600,
            "size": 16384,
            "index": 1
        },
        {
            "label": "Samsung DDR3 DIMM",
            "speed": 1600,
            "size": 16384,
            "index": 2
        },
        {
            "label": "Samsung DDR3 DIMM",
            "speed": 1600,
            "size": 16384,
            "index": 3
        },
        {
            "label": "Samsung DDR3 DIMM",
            "speed": 1600,
            "size": 16384,
            "index": 4
        }
    ]
}
```

As you can see, this structure is quite flat (and we will do our best to keep it
that way), consisting mostly of lists of dicts - a noticeable exception to this
"rule" is `mac_addresses`, but this is just temporary workaround (see next
section).

Keep in mind though, that this is just an initial version of contract, which are
subject to heavy changes until `ralph-cli` will reach `1.0.0` version.

### Output - draft of the next version

As mentioned previously, our Script Contract is going to evolve. Therefore,
below you will find a sneak-peek of its future shape (in a "draft" form,
i.e. subject to change).

```no-highlight
{
    "model_name": "Dell PowerEdge R620",
    "processors": [
        {
            "model_name": "Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz",
            "family": "B3", // to be removed
            "label": "Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz", // to be removed
            "index": 1, // to be removed
            "speed": 3600,
            "cores": 8
        }
    ],
    "fibre_channel_cards": [
        {
            "firmware": "",
            "model": "",
            "speed": "",
            "wwn": "",
            "pwwn": "" // ...and any other "wwn" that can be found (we need them all)
        }
    ],
    "ethernets": [
        {
            "mac_address": "AA:AA:AA:AA:AA:AA",
            "model": "Intel(R) Ethernet 10G 4P X520/I350 rNDC",
            "speed": "", // add if possible
            "firmware_version": "", // add if possible
        }
    ],
    "disks": [
        {
            "model_name": "ATA Samsung SSD 840",
            "family": "ATA", // to be removed
            "label": "ATA Samsung SSD 840", // to be removed
            "size": 476,
            "serial_number": "S1AXNSAD8000000",
            "slot": "", // add if possible
            "firmware_version": "" // add if possible
        },
    ],
    "memory": [
        {
            "label": "Samsung DDR3 DIMM", // should be renamed to "model_name"
            "speed": 1600,
            "size": 16384,
            "index": 1 // to be removed
        },
    ],
    "software": [
        {
            "type": "firmware", // possible types: "firmware" and "os"
            "name": "iLO2",
            "version": "1.77"
        }
    ]
}
```

If you have any thoughts on this (or if you need to add something here), please
let us know by opening a new issue on [our GitHub profile][issues].


[self-contract]: concepts.md#scripts-contract
[quickstart-further]: quickstart.md#going-further
[ideas]: development.md#ideas-for-future-development

[TOML]: https://github.com/toml-lang/toml
[virtualenv]: https://packaging.python.org/en/latest/installing/#creating-and-using-virtual-environments
[issues]: https://github.com/allegro/ralph-cli/issues
