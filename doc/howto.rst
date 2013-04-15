==========
How to use
==========

Use help command to display available options::

  $ ~/beast -h

otherwise...


List all available resources
----------------------------

If you want to preview all API resources: ::

  $ ~/beast show

Output: ::

  Ralph API, schema:
  --------------------------------------------------
  bladeserver
  businessline
  ci
  ...
  virtualserver
  windowsdevice


You can also list all the resource fields: ::

  $ ~/beast show venture --schema

Output: ::

  Ralph API > venture, schema
  --------------------------------------------------
  name*
  cache_version
  created
  symbol*
  modified
  devices
  roles
  show_in_ralph*
  department*
  path
  is_infrastructure
  id*
  resource_uri

If field name has ``*``, that you can execute filter on this filed


Show details of the selected resource
-------------------------------------
::

  $ ~/beast show venture

Output: ::

  ----------------------------------------------------------------
  | name      | cache_version| created   | symbol    | modified  |
  ----------------------------------------------------------------
  | Venture1  | 2            | 2012-06-13| venture1  | 2012-06-14|
  ----------------------------------------------------------------
  | Venture2  | 5            | 2012-01-23| venture2  | 2012-06-14|
  ----------------------------------------------------------------
  | Venture3  | 8            | 2011-10-12| venture3  | 2012-10-31|
  ----------------------------------------------------------------

Filter
~~~~~~

If you need to see filtered data use: ::

  $ ~/beast show venture --filter="symbol=venture2&modified=2012-06-14"

Available filters:
``exact``, ``iexact``, ``contains``, ``icontains``, ``in``, ``gt``, ``gte``, ``lt``, ``lte``, ``startswith``,
``istartswith``, ``endswith``, ``iendswith``, ``range``, ``year``, ``month``, ``day``, ``week_day``, ``hour``,
``minute``, ``second``, ``isnull``, ``search``, ``regex``, ``iregex``::

Read more about Django ORM filters_.

.. _filters: https://docs.djangoproject.com/en/dev/ref/models/querysets/#field-lookups

Output: ::

  -----------------------------------------------------------------------------------------------------
  | name   | cache_version| created   | symbol | modified  | devices| roles| show_in_ralph| department|
  -----------------------------------------------------------------------------------------------------
  | Venture2| 4           | 2012-01-23| venture2 | 2012-06-14||     |      | x            | 2         |



Fields
~~~~~~

If you need to filter data, and no builtin API filter is available, you can use
additional filtering by ordinar python expressions using ``row`` dict variable ::

  $ ~/beast show venture --fields="name, symbol"

Output: ::

  -------------------------
  | name      | symbol    |
  -------------------------
  | Venture1  | venture1  |
  -------------------------
  | Venture2  | venture2  |
  -------------------------
  | Venture3  | venture2  |
  -------------------------



Limit
~~~~~

Specifies the number of results::

  ~/beast show venture --limit=1

Output: ::

  Limited rows requested: 1
  ----------------------------------------------------------------
  | name      | cache_version| created   | symbol    | modified  |
  ----------------------------------------------------------------
  | Venture1  | 2            | 2012-06-13| venture1  | 2012-06-14|
  ----------------------------------------------------------------


Trim
~~~~

Use to better trim data::

  ~/beast show venture --trim


Width
~~~~~

Limit table width to the specified number of characters::

  ~/beast show venture --width=100


Debug
~~~~~

Shows request time::

  ~/beast show venture --debug


Export to the file
~~~~~~~~~~~~~~~~~~

Beast can prepare data to export ``csv`` or ``trim`` format.
::
  ~/beast export venture --csv > ~/ralph_ventures.csv

CSV file is encoding to ``Unicode(UTF-8)`` and separated by ``comma``.::
