# Overview

`ralph-cli` is a command-line interface for [Ralph][ralph]. Its goal is to serve
as a "Swiss Army knife" for all the Ralph's functionality that is reasonable
enough for bringing it from web GUI to your terminal. At this moment, you can
use it for discovering MAC addresses of your hardware (with `scan` command), but
we are going to extend the functionality in the future (see
[Ideas for Future Development][development-ideas]).

Please note that `ralph-cli` should be considered as "work in progress" aka
"early beta", so keep in mind that until the `1.0.0` version is reached, things
*will* get changed and *may* be broken!

## Relation to "beast" (old ralph-cli)

`ralph-cli`'s [GitHub repo][ralph-cli] used to be inhabited by `beast` - an
older version of Ralph's API command-line client, written in Python. It is no
longer maintained, but you can still find its code on the `beast` branch
(although don't be surprised if it will disappear some day).

## Where to go next?

You should start with [Quickstart](quickstart.md) to get up and running
quickly. You may want to consult [Key Concepts](concepts.md) section as well,
which is meant to be used as a more detailed reference.

[development-ideas]: development.md#ideas-for-future-development

[ralph]: https://github.com/allegro/ralph
[ralph-cli]: https://github.com/allegro/ralph-cli
