# Overview

`ralph-cli` is a command-line interface for [Ralph][ralph]. At this moment, its
functionality is pretty basic (not to say scanty) - only `scan` command, which
works only for network cards (by detecting MAC addresses on a given host and
sending them to Ralph), but eventually, `ralph-cli` should serve as a "Swiss
Army knife" for all the Ralph's functionality that is reasonable enough for
bringing it from web GUI to your terminal (deployment, anyone..?).

`ralph-cli` should be considered as "work in progress" aka "early beta", so keep
in mind that until the `1.0.0` version is reached, things *will* get changed and
*may* be broken!

## Relation to "beast" (old ralph-cli)

`ralph-cli`'s [GitHub repo][ralph-cli] used to be inhabited by `beast` - an
older version of Ralph's API command-line client, written in Python. It is no
longer maintained, but you can still find its code on the `beast` branch
(although don't be surprised if it will disappear some day).

## Where to go next?

You should start with [Quickstart](quickstart.md) to get up and running
quickly. You may want to consult [Key Concepts](concepts.md) section as well,
which is meant to be used as a more detailed reference. Finally, you may want to
check [Ideas for Future Development](ideas.md), which should give you an
approximate picture in which direction `ralph-cli`'s development is heading.

[ralph]: https://github.com/allegro/ralph
[ralph-cli]: https://github.com/allegro/ralph-cli
