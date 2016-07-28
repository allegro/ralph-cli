# Change Log

## 0.2.0

Released on July 28, 2016.

* `scan` command: ability to handle `Ethernet`, `Memory`, `FibreChannelCard`,
  `Processor` and `Disk` components as well as `FirmwareVersion` and
  `BIOSVersion` fields on `DataCenterAsset`.

## 0.1.0

Released on June 29, 2016.

* Ability to handle Python virtual environments.
* Initial version of manifest files for scan scripts.
* Ability to configure `ralph-cli` via config file.
* Initial versions of two scanning scripts (`idrac.py` and `ilo.py`).
* Integration with [Travis][] and [Coveralls][].
* Initial version of `scan` command - only detected MAC addresses are sent to
  Ralph.

[Travis]: https://travis-ci.org/
[Coveralls]: https://coveralls.io/
