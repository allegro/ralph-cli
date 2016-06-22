# Ideas for Future Development

This section collects some of the ideas that may be implemented in
`ralph-cli`. It should give you an approximate picture in which direction
`ralph-cli`'s development is heading.

## Sooner

* Ability to refresh/recreate virtualenvs used by scan scripts written in Python
  (e.g. after adding new dependency to manifest file).
* Integration with [Logstash][logstash] (this is almost ready, though).
* Ability to create/update/delete other components than network cards,
  e.g. `ralph-cli scan --components=disk,memory,processor` etc.
* Ability to feed `ralph-cli scan` with ready-made JSON files (i.e. without
  launching any scan scripts).
* Ability to update all components detected by scan on a given host at once
  (i.e. with a single HTTP request over a single API endpoint).
* Ability to configure `ralph-cli` by environment variables, which would take
  precedence over the config file.
* Ability to scan whole range of hosts/networks (at this moment, `ralph-cli`
  operates only on single hosts).
* Some minor improvements like setting timeouts for scan, adding progress bars etc.

## Later

* Ability to add/checkout scan scripts to/from a Git repo without direct
  interaction with Git itself, e.g. `ralph-cli script --commit`,
  `ralph-cli script --checkout` - this should include commands like
  `ralph-cli script --edit` and so on.
* Ability to deploy hosts without touching Ralph's GUI, e.g. `ralph-cli deploy`.

[logstash]: https://www.elastic.co/products/logstash
