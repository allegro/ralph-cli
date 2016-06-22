# ralph-cli

[![Build Status](https://travis-ci.org/allegro/ralph-cli.svg?branch=develop)](https://travis-ci.org/allegro/ralph-cli)
[![Coverage Status](https://coveralls.io/repos/github/allegro/ralph-cli/badge.svg?branch=develop)](https://coveralls.io/github/allegro/ralph-cli?branch=develop)
[![Documentation Status](https://readthedocs.org/projects/ralph-cli/badge/?version=latest)](http://ralph-cli.readthedocs.io/en/latest/?badge=latest)

`ralph-cli` is a command-line interface for [Ralph][ralph]. At this moment, its
functionality is pretty basic (not to say scanty) - only `scan` command, which
works only for network cards, by detecting MAC addresses on a given host and
sending them to Ralph - but eventually, `ralph-cli` should serve as a "swiss
army knife" for all the Ralph's functionality that is reasonable enough for
bringing it from web GUI to your terminal (deployment, anyone..?).

`ralph-cli` should be considered as "work in progress" aka "early beta", so keep
in mind that until the `1.0.0` version is reached, things *will* get changed and
*may* be broken!

## Relation to "beast" (old ralph-cli)

`ralph-cli`'s repo used to be inhabited by `beast` - an older version of Ralph's
API command-line client, written in Python. It is no longer maintained, but you
can still find its code on the `beast` branch (although don't be surprised if it
will dissappear some day).

## Where to go next?

You should start with [Quickstart][quickstart] to get up & running
quickly. You may want to consult [Key Concepts][concepts] section as well,
which is meant to use as a more detailed reference. Finally, you may want to
check [Ideas for Future Development][ideas], which should give you an
approximate picture in which direction `ralph-cli`'s development is heading.

## License

`ralph-cli` is licensed under the [Apache License, v2.0][apache]. Copyright (c)
2016 [Allegro Group][allegro]

[glide]: https://github.com/Masterminds/glide
[apache]: http://www.apache.org/licenses/LICENSE-2.0
[allegro]: http://allegrogroup.com
[quickstart]: http://ralph-cli.readthedocs.io/en/latest/quickstart/
[concepts]: http://ralph-cli.readthedocs.io/en/latest/concepts/
[ideas]: http://ralph-cli.readthedocs.io/en/latest/ideas/
