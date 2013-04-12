#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""Beast is convenient Ralph API commandline client.
Documentation on https://github.com/allegro/ralph_beast/
Example usage $ ~/beast show

Usage:
 beast show
 beast show <resource> [--debug] [--schema] [--fields=fields] [--filter=field_filter] [--limit=limit] [--csv] [--trim] [--width=max_width] [--silent]
 beast update <resource> <id> <fields> <fields_values>
 beast -h | --help
 beast --version

"""

import platform
PLATFORM = platform.system()
if not PLATFORM == 'Windows':
    import fcntl  # pdcurses
    import termios

import codecs
import cStringIO
import csv
import errno
import os
import requests
import slumber
import struct
import sys
import time
import urlparse

from collections import defaultdict
from docopt import docopt

SHOW_VERBOSE = True


def llen(s):
    """Return length of utf-8 encoded string"""
    return len(s) + 4


class Api(object):
    def get_session(self, settings,):
        url = settings.get('url')
        if url.endswith('/'):
            url = url.rstrip('/')
        version = settings.get('version')
        session = requests.session(verify=False)
        session = slumber.API(
            '%(url)s/api/v%(version)s/' % dict(url=url, version=version),
            session=session,
        )
        return session

    def put_resource(self, settings, resource, id, data,):
        session = self.get_session(settings)
        username = settings.get('username')
        api_key = settings.get('api_key')
        try:
            data = getattr(session, resource)(id).put(
                data=data,
                username=username,
                api_key=api_key,
            )
        except slumber.exceptions.HttpClientError as error:
            print_err('Error: ', error.content)
        return data

    def get_resource(self, settings, resource, limit=None, filters=None,):
        session = self.get_session(settings)
        resource = '' if not resource else resource
        limit = 0 if not limit else limit
        attrs = [
            ('username', settings.get('username')),
            ('api_key', settings.get('api_key')),
            ('limit', limit),
        ]
        if filters:
            url_dict = urlparse.parse_qsl(filters)
            attrs.extend(url_dict)
        attrs_dict = dict(attrs)
        try:
            return getattr(session, resource).get(**dict(attrs_dict))
        except slumber.exceptions.HttpClientError as e:
            if e.response.status_code == 401:
                print_err("Error: Authorization failed")
                sys.exit(9)
            else:
                print_err('Error: %s' % e.content)
                sys.exit(10)
        except slumber.exceptions.HttpServerError as e:
            print_err('Server error: %s' % e.content)
            sys.exit(10)


    def get_schema(self, settings, resource=None, filters=False,):
        rows, columns = get_terminal_size()
        if not resource:
            print("-" * columns)
            data = self.get_resource(settings, '')
            list_of_resources = [api_resource for api_resource in data]
            print('\n'.join(sorted(list_of_resources)))
        else:
            print("-" * columns)
            schema = '%s/schema/' % resource
            data = self.get_resource(settings, resource=schema)
            for field_name in data['fields']:
                field_dict = data['fields'][field_name]
                if field_name in data.get('filtering', {}).keys():
                    mark = '*'
                else:
                    mark = ''
                print(
                    '%(mark)s%(field_name)s - %(help)s' % dict(
                        mark=mark,
                        field_name=field_name,
                        help=field_dict.get('help_text'),
                    )
                )


class Content(object):
    def __init__(self, api_data, fields_requested):
        self.api_data = api_data
        self.fields_requested = fields_requested

    def get_repr_rows(self):
        return [self.row_repr(row) for row in self.api_data]

    def field_repr(self, field):
        rep = ''
        if isinstance(field, basestring):
            if '/api/v0.9/' in field:
                rep = unicode(field.replace('/api/v0.9/', ''))
            else:
                rep = unicode(field[:20])
        elif isinstance(field, bool):
            rep = 'x' if field else 'o'
        elif isinstance(field, dict):
            try:
                # choice
                rep = unicode(field['id'])
            except KeyError:
                rep = unicode(field[field.keys()[0]])
        elif field is None:
            rep = ''
        elif isinstance(field, list):
            rep = ','.join(
                [self.field_repr(subfield) for subfield in field]
            )
        else:
            rep = unicode(field)
        return rep

    def row_repr(self, row,):
        for field in row.keys():
            encode_field = field.encode('utf-8')
            row[encode_field] = self.field_repr(row[encode_field])
        return row


def __get_terminal_size_windows():
    res = None
    try:
        from ctypes import windll, create_string_buffer
        h = windll.kernel32.GetStdHandle(-12)
        csbi = create_string_buffer(22)
        res = windll.kernel32.GetConsoleScreenBufferInfo(h, csbi)
    except:
        return 80, 25
    if res:
        (bufx, bufy, curx, cury, wattr,
         left, top, right, bottom, maxx, maxy) = struct.unpack("hhhhHhhhhhh", csbi.raw)
        return bufy, bufx
    else:
        return 80, 25


def get_terminal_size():
    if PLATFORM == 'Windows':
        return __get_terminal_size_windows()
    elif sys.stdout.isatty():
        data = fcntl.ioctl(sys.stdout.fileno(), termios.TIOCGWINSZ, '1234')
        return struct.unpack('hh', data)
    return 60, 120


class Writer(object):
    data = []

    def write_header(self, *args, **kwargs):
        raise NotImplementedError()

    def write_row(self, *args, **kwargs):
        raise NotImplementedError()

    def get_all_columns(self):
        for row in self.data:
            return row.keys()
        else:
            return []


class ConsoleWriter(Writer):
    def __init__(self, data, columns_requested, trim, max_width):
        self.data = data
        self.columns_requested = columns_requested
        self.max_width = max_width
        self.columns_widths = defaultdict(int)
        self.columns_truncated = []
        self.calculate_columns_widths(trim)
        self.columns_visible = self.get_visible_columns()

    def get_visible_columns(self):
        columns_visible = []
        current_width = 0
        for key in self.columns_requested or self.get_all_columns():
            width = self.columns_widths.get(key)
            if current_width + width > self.max_width:
                return columns_visible
            else:
                columns_visible.append(key)
            current_width += width
        return columns_visible

    def calculate_columns_widths(self, trim):
        data_to_measure = [{key: key} for key in self.get_all_columns()]
        if not trim:
            data_to_measure += self.data

        for row in data_to_measure:
            for key in row.keys():
                self.columns_widths[key] = max(self.columns_widths[key], llen(row[key]))

    def write_header(self):
        print("-" * self.max_width)
        sys.stdout.write("|")
        for key in self.columns_visible:
            fill = self.columns_widths.get(key)
            sys.stdout.write(
                ' {:<{fill}} |'.format(
                    key[:fill-4].encode('utf-8'),
                    fill=fill-4,
                )
            )
        sys.stdout.write('\n')
        print("-" * self.max_width)

    def write_row(self, row):
        sys.stdout.write("|")
        for key in self.columns_visible:
            fill = self.columns_widths.get(key)
            sys.stdout.write(
                ' {:<{fill}} |'.format(
                    row[key][:fill-4].encode('utf-8'),
                    fill=fill-4,
                )
            )
        sys.stdout.write('\n')


class CSVWriter(Writer):
    def __init__(self, data, columns_requested, trim, max_width):
        self.data = data
        self.columns_requested = columns_requested
        self.trim = trim
        self.max_width = max_width
        f = cStringIO.StringIO()
        dialect = csv.excel
        encoding = "utf-8"
        self.queue = cStringIO.StringIO()
        self.writer = csv.writer(self.queue, dialect=dialect)
        self.stream = f
        self.encoder = codecs.getincrementalencoder(encoding)()

    def _write(self, row):
        self.writer.writerow([item.encode("utf-8") for item in row])
        data = self.queue.getvalue()
        data = data.decode("utf-8")
        data = self.encoder.encode(data)
        self.stream.write(data)
        self.queue.truncate(0)
        return data

    def write_header(self):
        sys.stdout.write(self._write(
            self.columns_requested or self.get_all_columns()))

    def write_row(self, row):
        columns = self.columns_requested or self.get_all_columns()
        row = ([value for key, value in row.iteritems() if key in columns])
        sys.stdout.write(self._write(row))


def show(arguments, settings):
    silent = arguments.get('--silent')
    if silent:
        global SHOW_VERBOSE
        SHOW_VERBOSE = False

    resource = arguments.get('<resource>', '')
    limit = arguments.get('--limit')
    if not limit and sys.stdout.isatty():
        # default limit for console is 20
        limit = 20

    fields_requested = arguments.get('--fields')
    fields_requested = [
        field.strip() for field in fields_requested.split(',')
    ] if fields_requested else []
    if not resource:
        print_debug("Available resources:")
        return Api().get_schema(settings, None)
    elif arguments.get('--schema'):
        print_debug("Schema of `%s` resource:" % resource)
        return Api().get_schema(settings, resource)

    print_debug("Resource: `%s` " % resource)
    response = Api().get_resource(
        settings,
        resource,
        limit,
        arguments.get('--filter'),
    )

    total_count = response['meta']['total_count']
    limit = response['meta']['limit']
    #next_link = response['meta'].get('next')

    api_data = response.get('objects', [])
    content = Content(api_data, fields_requested)
    rows, columns = get_terminal_size()
    max_width = int(arguments.get('--width') or columns)

    print_debug("Total count: %s" % total_count)
    if limit:
        print_debug("Limit: %s" % limit)

    trim = arguments.get('--trim')
    parameters = dict(data=content.get_repr_rows(), columns_requested=fields_requested,
                      trim=trim, max_width=max_width)
    writer_class = CSVWriter if arguments.get('--csv') else ConsoleWriter
    writer = writer_class(**parameters)
    writer.write_header()
    for row in api_data:
        writer.write_row(row)


def update(arguments, settings,):
    resource = arguments.get('<resource>')
    id = arguments.get('<id>')
    fields = arguments.get('<fields>').split(',')
    fields_values = arguments.get('<fields_values>')
    for row in csv.reader(
        [fields_values],
        quotechar='"', quoting=csv.QUOTE_MINIMAL
    ):
        pass
    data = dict(zip(fields, row))
    Api().put_resource(settings, resource, id, data)


def do_main(arguments,):
    debug = arguments.get('--debug')
    if debug:
        stopwatch_start = time.time()
    settings = dict()
    config_file = os.path.abspath("config")
    additional_config_file = os.path.expanduser("~/.beast/config")
    try:
        execfile(config_file, settings)
    except IOError:
        try:
            execfile(additional_config_file, settings)
        except IOError, e:
            if e.errno == errno.ENOENT:
                print_err("Config file doesn't exist.")
            else:
                print_err(e)
            sys.exit(4)
    required_keys = ('url', 'username', 'api_key', 'version')
    missing_options = set(required_keys) - set(settings)
    if missing_options:
        print_err(
            'Config file options: %s are required.' %
            ','.join(missing_options)
        )
        sys.exit(5)

    if arguments.get('show'):
        show(arguments, settings)
    elif arguments.get('update'):
        update(arguments, settings)

    if debug:
        print_debug('\nTotal time: %s sec' % round(time.time()-stopwatch_start, 2))


def main():
    arguments = docopt(__doc__, version='1.2.3')
    do_main(arguments)


def print_debug(s):
    if SHOW_VERBOSE:
        sys.stderr.write(s + '\n')


def print_err(s):
    sys.stderr.write(s + '\n')

if __name__ == '__main__':
    main()
