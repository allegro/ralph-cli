The following HOWTO assumes that you've successfully installed and configured ralph ralph-cli according to the instructions available `here <./install.rst>`_.

==========
How to use
==========

Use help command to display available options::

  $ ~/ralph-cli -h

otherwise...


List all available resources
----------------------------

If you want to preview all API resources: ::

  $ ~/ralph-cli list

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

  $ ~/ralph-cli data-center-assets --schema

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

  $ ~/ralph-cli data-center-assets list

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

  $ ~/ralph-cli data-center-assets list --filter="modified=2012-06-14&symbol__in=venture1,venture2"

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
  | Venture1| 2           | 2012-06-13| venture1 | 2012-06-14||     |      |              | 2         |
  -----------------------------------------------------------------------------------------------------
  | Venture2| 4           | 2012-01-23| venture2 | 2012-06-14||     |      | x            | 2         |



Fields
~~~~~~

If you need to filter data, and no builtin API filter is available, you can use
additional filtering by ordinar python expressions using ``row`` dict variable ::

  $ ~/ralph-cli venture list --fields="name, symbol"

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


Subfields (dicts)
~~~~~~~~~~~~~~~~~~~~

When the return data consists of mulptiple fields you should decide which field to display. If you don't do this
generic 'dict' will be returned.

An example is `ip_addresses` field of `dev` resource. Here you should point which subfield to display. ::

  $ ~/ralph-cli show dev --fields="name, ip_addresses"

  -------------------------------------------------------------------------------------------------------------------
  | ip_addresses                                                          | name             |
  -------------------------------------------------------------------------------------------------------------------
  | dict#dict#dict                                                        | test.testx       |
  | dict                                                                  | Rack 105         |

test.testx has 3 ip_addresses which consists of subfields.

Specify subfield with `field:subfield` statement. You can inspect subfields by specifying `:?`
  
Example: Examine all available subfields for `ip_addresses` ::

  $ ~/ralph-cli show dev --fields="name, ip_addresses:?"

  Available keys: snmp_community,snmp_version,number,network,network_details,created,hostname,last_plugins,modified,is_management,http_family,dead_ping_count,is_buried,last_puppet,address,device,is_public,resource_uri,id,last_seen19


Now just specify `address` subfield and export csv ::

  $ ralph-cli show dev --fields=ip_addresses:address --csv

  ip_addresses,name
  "10.10.10.10,5.5.5.5",hostname.dc3
  "10.10.10.3",hostname.dc4

Beware: Currently pretty printed tabular output for subfields is not supported - use csv export instead.

Limit
~~~~~

Specifies the number of results::

  ~/ralph-cli show venture --limit=1

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

  ~/ralph-cli show venture --trim


Width
~~~~~

Limit table width to the specified number of characters::

  ~/ralph-cli show venture --width=100


Debug
~~~~~

Shows request time::

  ~/ralph-cli show venture --debug


Export to the file
~~~~~~~~~~~~~~~~~~

Beast can export to the ``csv`` format. ::

  ~/ralph-cli show venture --csv > ~/ralph_ventures.csv

CSV file is encoding to ``Unicode(UTF-8)`` and separated by ``comma``.::


Add resource
----------------------------

If you want to create new object through the API use following statement ::

  $ ~/ralph-cli create [resource] [--file=file_path] [--data='json_data']

Some of the fields are required for given Resource - field names are identical
with `ralph-cli show` output. ::


  $ cat /tmp/data.json

  {
        "status" : 2,
        "name" : "some.ci.name",
        "technical_owners": [],
        "business_owners": [],
        "layers" : [
          {
            "name" : "Hardware"
          }
        ],
        "type" : {
          "name" : "Device"
        },
        "state" : 2,
        "barcode" : "come.unique.barcode"
  }

You can use - file for stding as well: ::

  $ cat /tmp/data.json | ~/ralph-cli create ci --file=-

Or specify data explicit in commandline: ::

  $ ~/ralph-cli create ci --data='{ "status" : 2, "name": "some.ci.name", ... }'
 

Update resource
---------------

If you want to update resource use following statement ::

  $ ~/ralph-cli update [resource] [id] [field1],[field2] [value1],[value2] 


Example ::

  $ ralph-cli update ci 1 name new_name


For data security reasons you can update only 1 resource at once - use multiple 
ralph-cli update invocations in shell scripts for bulk changes.

