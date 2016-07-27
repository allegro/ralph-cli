# ralph-cli

[![Build Status](https://travis-ci.org/allegro/ralph-cli.svg?branch=develop)](https://travis-ci.org/allegro/ralph-cli)
[![Coverage Status](https://coveralls.io/repos/github/allegro/ralph-cli/badge.svg?branch=develop)](https://coveralls.io/github/allegro/ralph-cli?branch=develop)
[![Go Report Card](https://goreportcard.com/badge/github.com/allegro/ralph-cli)](https://goreportcard.com/report/github.com/allegro/ralph-cli) 
[![Documentation Status](https://readthedocs.org/projects/ralph-cli/badge/?version=latest)](http://ralph-cli.readthedocs.io/en/latest/?badge=latest)

`ralph-cli` is a command-line interface for [Ralph][ralph]. Its goal is to serve
as a "Swiss Army knife" for all the Ralph's functionality that is reasonable
enough for bringing it from web GUI to your terminal. At this moment, you can
use it for discovering components of your hardware (with `scan` command), but we
are going to extend the functionality in the future (see
[Ideas for Future Development][development-ideas]).

Please note that `ralph-cli` should be considered as "work in progress" aka
"early beta", so keep in mind that until the `1.0.0` version is reached, things
*will* get changed and *may* be broken!

At this moment, we support only Linux and Mac OS X operating systems, but we
have Windows on our roadmap, so stay tuned.

## Relation to "beast" (old ralph-cli)

`ralph-cli`'s [GitHub repo][ralph-cli] used to be inhabited by `beast` - an
older version of Ralph's API command-line client, written in Python. It is no
longer maintained, but you can still find its code on the `beast` branch
(although don't be surprised if it will disappear some day).

## Where to go next?

You should start with [Quickstart][quickstart] to get up and running
quickly. You may want to consult [Key Concepts][concepts] section as well,
which is meant to be used as a more detailed reference.


## License

`ralph-cli` is licensed under the [Apache License, v2.0][apache]. Copyright (c)
2016 [Allegro Group][allegro]

[development-ideas]: http://ralph-cli.readthedocs.io/en/latest/development/#ideas-for-future-development
[quickstart]: http://ralph-cli.readthedocs.io/en/latest/quickstart/
[concepts]: http://ralph-cli.readthedocs.io/en/latest/concepts/

[ralph]: https://github.com/allegro/ralph
[ralph-cli]: https://github.com/allegro/ralph-cli
[glide]: https://github.com/Masterminds/glide
[apache]: http://www.apache.org/licenses/LICENSE-2.0
[allegro]: http://allegrogroup.com
