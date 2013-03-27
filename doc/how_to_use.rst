==========
How to use
==========

Use help command to display available options::

  $ ~/beast -h

otherwise...


List all available resources
----------------------------

If you want to preview all api resources: ::

  $ ~/beast inspect

Output: ::

  Available resources:
  --------------------------------------------------
  bladeserver
  businessline
  ci
  ...
  virtualserver
  windowsdevice


You can also list all the resource fields: ::

  $ ~/beast inspect --resource=venture

Output: ::

  Available fields for resource: venture
  --------------------------------------------------
  cache_version
  created
  department
  devices
  id
  is_infrastructure
  modified
  name
  path
  resource_uri
  roles
  show_in_ralph
  symbol


Show details of the selected resource
-------------------------------------
::

  $ ~/beast export venture

Output: ::

  ------------------------------------------------------------------------------------------------------------------------
  | name      | cache_version| created   | symbol    | modified  |
  ------------------------------------------------------------------------------------------------------------------------
  | Venture1  | 2            | 2012-06-13| venture1  | 2012-06-14|
  ------------------------------------------------------------------------------------------------------------------------
  | Venture2  | 5            | 2012-01-23| venture2  | 2012-06-14|
  ------------------------------------------------------------------------------------------------------------------------
  | Venture3  | 8            | 2011-10-12| venture3  | 2012-10-31|
  ------------------------------------------------------------------------------------------------------------------------

Filter
~~~~~~

If you need to see filtered data you can use Python::

  $ ~/beast export venture --filter="row.get('symbol') == 'venture2'"

Output: ::

  ------------------------------------------------------------------------------------------------------------------------
  | name   | cache_version| created   | symbol | modified  | devices| roles| show_in_ralph| department|
  ------------------------------------------------------------------------------------------------------------------------
  | Venture2| 4            | 2012-01-23| venture2 | 2012-06-14||        |      | x            | 2         |



Fields
~~~~~~

If you need to filter data, and no builtin API filter is available, you can use
additional filtering by ordinar python expressions using ``row`` dict variable ::

  $ ~/beast export venture --fields=name,symbol

Output: ::

  ------------------------------------------------------------------------------------------------------------------------
  | name      | symbol    |
  ------------------------------------------------------------------------------------------------------------------------
  | Venture1  | venture1  |
  ------------------------------------------------------------------------------------------------------------------------
  | Venture2  | venture2  |
  ------------------------------------------------------------------------------------------------------------------------
  | Venture3  | venture2  |
  ------------------------------------------------------------------------------------------------------------------------




Limit
~~~~~

Specifies the number of results::

  ~/beast export venture --limit=1

Output: ::

  Limited rows requested: 1
  ------------------------------------------------------------------------------------------------------------------------
  | name      | cache_version| created   | symbol    | modified  |
  ------------------------------------------------------------------------------------------------------------------------
  | Venture1  | 2            | 2012-06-13| venture1  | 2012-06-14|
  ------------------------------------------------------------------------------------------------------------------------


Export to the file
~~~~~~~~~~~~~~~~~~

Beast can prepare data to export ``csv``, ``yaml`` or ``trim`` format.
::
  ~/beast export venture --csv > ~/ralph_ventures.csv

If you use Windows, yours home directory path is: ::

  c:\cygwin\home\windows_user_name\

You can also open file from console in yours text editor and save on preferred
place. ::

  cygstart.exe ~/ralph_ventuures.csv

