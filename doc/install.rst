==============
How to install
==============



Windows
-------

Download `beast.exe` file and prepare the configuration file.

.. _beast.exe: https://github.com/allegro/ralph_beast/raw/master/beast_windows.zip

The configuration file should be placed in the same directory as ``beast.exe`` or home directory.
Config file should be contain such data  :ref:`config_file`


MacOS
---------------

Put on your console below command to install.::

  $ pip install ralph-beast

or: ::

  $ curl https://raw.github.com/allegro/ralph_beast/master/install.sh | bash -



Linux
---------------

Please install curl if not installed, and identically to MacOS install:

  $ pip install ralph-beast

or: ::

  $ curl https://raw.github.com/allegro/ralph_beast/master/install.sh | bash -


Now do you need configuration file. In your home directory create directory
``.beast`` and add text file named ``config``.

Config file should be contain such data  :ref:`config_file`


.. _config_file:

Config file - example
---------------------
::

  username="jan.kowalski"
  api_key="478457f9f32323201ebde8ef79cd9d3a028ced56747"
  url="https://ralph-url.com"
  version="0.9"
