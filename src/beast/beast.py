#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""Beast is convenient Ralph API commandline client.
Documentation on https://github.com/allegro/ralph_beast/
Example usage $ ~/beast show

Usage:
 beast show
 beast show <resource> [--debug] [--schema] [--fields=fields] [--filter=field_filter] [--csv] [--trim] [--width=max_width] [--silent] [--limit=limit]
 beast update <resource> <id> <fields> <fields_values>
 beast create <resource> [--data=json_data] [--file=path_to_the_json_file]
 beast -h | --help
 beast --version

"""

import platform
PLATFORM = platform.system()
if not PLATFORM == 'Windows':
    import fcntl  # pdcurses
    import termios

import cStringIO
import codecs
from collections import defaultdict
from collections import OrderedDict
import csv
import json
import os
import struct
import sys
import time
import urlparse

from colorconsole import terminal
from docopt import docopt
import errno
import requests
from requests.auth import AuthBase
import slumber

from . import VERSION

SHOW_VERBOSE = True


class ErrorHandlerContext(object):

    def __exit__(self, type_, value, exception):
        if type_ is slumber.exceptions.HttpServerError:
            msg = json.loads(value.content)['error_message']
            print_err('Error: %s' % msg)
            ret = 3
        elif type_ is slumber.exceptions.HttpClientError:
            ret = 2
            print_err('Error: %s ' % value)
        elif type_ is None:
            return
        else:
            ret = 4
            print_err('Other Error: %s' % value)
        sys.exit(ret)

    def __enter__(self):
        pass


class TastypieApikeyAuth(AuthBase):

    def __init__(self, username, apikey):
        self.username = username
        self.apikey = apikey

    def __call__(self, r):
        r.headers['Authorization'] = "ApiKey {0}:{1}".format(
            self.username, self.apikey)
        return r


def llen(s):
    """Return length of utf-8 encoded string"""
    return len(s) + 4


class Api(object):

    def get_session(self, settings,):
        url = settings.get('url')
        if url.endswith('/'):
            url = url.rstrip('/')
        version = settings.get('version')
        username = settings.get('username')
        api_key = settings.get('api_key')
        session = requests.session(verify=False)
        session.auth = TastypieApikeyAuth(username, api_key)
        session = slumber.API(
            '%(url)s/api/v%(version)s/' % dict(url=url, version=version),
            session=session,
        )
        return session

    def create_resource(self, settings, resource, data):
        session = self.get_session(settings)
        with ErrorHandlerContext():
            data = getattr(session, resource).post(
                data=data,
            )
        return data

    def patch_resource(self, settings, resource, id, data):
        session = self.get_session(settings)
        with ErrorHandlerContext():
            data = getattr(session, resource)(id).patch(
                data=data,
            )
        return data

    def get_resource(self, settings, resource, limit=20, offset=0, filters=None):
        session = self.get_session(settings)
        resource = '' if not resource else resource
        attrs = [
            ('limit', limit),
            ('offset', offset),
        ]
        if filters:
            url_dict = urlparse.parse_qsl(filters)
            attrs.extend(url_dict)
        attrs_dict = dict(attrs)
        with ErrorHandlerContext():
            return getattr(session, resource).get(**dict(attrs_dict))

    def get_schema(self, settings, resource=None, filters=False,):
        rows, columns = get_terminal_size()
        if not resource:
            cprint("-" * columns, 'GREEN')
            data = self.get_resource(settings, '')
            list_of_resources = [api_resource for api_resource in data]
            print('\n'.join(sorted(list_of_resources)))
        else:
            cprint("-" * columns, 'GREEN')
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

    def field_repr(self, field, field_name):
        rep = ''
        if isinstance(field, basestring):
            if '/api/v0.9/' in field:
                rep = unicode(field.replace('/api/v0.9/', ''))
            else:
                rep = unicode(field)
        elif isinstance(field, bool):
            rep = '1' if field else '0'
        elif isinstance(field, dict):
            path_requested = self.fields_requested.get(field_name)
            subdict = field
            if path_requested:
                for chain in path_requested:
                    if chain == '?':
                        if not isinstance(subdict, dict):
                            cprint('\nError - subobject is not dictionary\n')
                            sys.exit(5)
                        cprint('\nAvailable keys: %s' %
                               ','.join(subdict.keys()))
                        sys.exit(4)
                    subdict = subdict[chain]
                return self.field_repr(subdict, field_name)
            return 'dict'
        elif field is None:
            rep = ''
        elif isinstance(field, list):
            rep = '#'.join(
                [self.field_repr(subfield, field_name) for subfield in field]
            )
        else:
            rep = unicode(field)
        return rep

    def row_repr(self, row):
        for field in row.keys():
            encode_field = field.encode('utf-8')
            row[encode_field] = self.field_repr(
                row[encode_field], encode_field)
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
        for key in (self.columns_requested or self.get_all_columns()):
            width = self.columns_widths.get(key) or 0
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
                self.columns_widths[key] = max(
                    self.columns_widths[key], llen(row[key]))

    def write_header(self):
        end_line = ''
        if not PLATFORM == 'Windows':
            end_line = '\n'
        cprint("-" * self.max_width + end_line, 'GREEN')
        cprint("|", 'GREEN')
        for key in self.columns_visible:
            fill = self.columns_widths.get(key) or 4
            cprint(
                ' {:<{fill}}'.format(
                    key[:fill - 4].encode('utf-8'),
                    fill=fill - 4,
                ),
                'LGREEN',
            )
            cprint(" |", 'GREEN')
        cprint('\n')
        cprint("-" * self.max_width + end_line, 'GREEN')

    def write_row(self, row):
        cprint("|", 'GREEN')
        for key in self.columns_visible:
            fill = self.columns_widths.get(key) or 4
            cprint(
                ' {:<{fill}}'.format(
                    unicode(row[key])[:fill - 4].encode('utf-8', 'ignore'),
                    fill=fill - 4,
                ),
                'WHITE',
            )
            cprint(" |", 'GREEN')
        cprint('\n')


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
        self.writer.writerow(
            [item.encode("utf-8", "ignore") if not isinstance(item, int) else unicode(item) for item in row])
        data = self.queue.getvalue()
        data = data.decode("utf-8")
        data = self.encoder.encode(data)
        self.stream.write(data)
        self.queue.truncate(0)
        return data

    def write_header(self):
        cprint(self._write(
            self.columns_requested or self.get_all_columns()))

    def write_row(self, row):
        columns = self.columns_requested or self.get_all_columns()
        row_p = [row[key] for key in columns]
        cprint(self._write(row_p))


def show(arguments, settings):
    silent = arguments.get('--silent')
    if silent:
        global SHOW_VERBOSE
        SHOW_VERBOSE = False

    resource = arguments.get('<resource>', '')
    limit_requested = arguments.get('--limit')
    if sys.stdout.isatty() and (limit_requested is None):
        # default limit for console is 20
        limit_requested = 20
    elif limit_requested is None:
        # file output
        limit_requested = 0

    fields_r = arguments.get('--fields')
    fields_requested_args = fields_r.split(',') if fields_r else []
    fields_requested = OrderedDict()
    for f in fields_requested_args:
        field = f.strip().lower()
        key = field
        if ':' in field:
            key = field.split(':')[0]
            full_path = field.split(':')[1:]
            fields_requested[key] = full_path
        else:
            fields_requested[key] = []

    if not resource:
        cprint("Available resources:\n", 'LCYAN', verbose=True)
        return Api().get_schema(settings, None)
    elif arguments.get('--schema'):
        cprint("Schema of `%s` resource:\n" % resource, 'LCYAN', verbose=True)
        return Api().get_schema(settings, resource)

    cprint("Resource: `%s` \n" % resource, 'LCYAN', verbose=True)
    finished = False

    first = True
    offset = 0
    fetched = 0
    while not finished:
        response = Api().get_resource(
            settings,
            resource,
            limit=50,
            offset=offset,
            filters=arguments.get('--filter'),
        )
        total_count = response['meta']['total_count']
        limit = response['meta']['limit']
        next_link = response['meta'].get('next')
        if not first and not next_link:
            finished = True
            break
        offset += (int(limit) + 1)

        api_data = response.get('objects', [])
        content = Content(api_data, fields_requested)
        rows, columns = get_terminal_size()
        max_width = int(arguments.get('--width') or columns)
        trim = arguments.get('--trim')

        if first:
            cprint("Total count: %s\n" % total_count, 'LCYAN', verbose=True)
            if limit:
                cprint("Limit: %s\n" % limit_requested, 'LCYAN', verbose=True)

        parameters = dict(data=content.get_repr_rows(), columns_requested=fields_requested,
                          trim=trim, max_width=max_width)
        writer_class = CSVWriter if arguments.get('--csv') else ConsoleWriter
        writer = writer_class(**parameters)
        if first:
            writer.write_header()
            first = False
        for row in api_data:
            fetched += 1
            if limit_requested and (fetched > limit_requested):
                finished = True
                break
            try:
                writer.write_row(row)
            except KeyError as e:
                print_err(
                    "\n\rCan't find column: %s. Type bast show [resource] --schema to show available columns." % e.args[0])
                sys.exit(2)


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
    Api().patch_resource(settings, resource, id, data)


def create(arguments, settings,):
    resource = arguments.get('<resource>')
    data = arguments.get('--data')
    if arguments.get('--file'):
        fname = arguments.get('--file')
        if fname == '-':
            data = sys.stdin.read().strip()
        else:
            data = open(fname).read().strip()
    Api().create_resource(settings, resource, data=json.loads(data))


def do_main(arguments,):
    debug = arguments.get('--debug')
    if debug and SHOW_VERBOSE:
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
    elif arguments.get('create'):
        create(arguments, settings)
    if debug and SHOW_VERBOSE:
        cprint(
            '\nTotal time: %s sec\n' % round(time.time() - stopwatch_start, 2),
            'LCYAN',
            verbose=True,
        )


def main():
    arguments = docopt(__doc__, version='.'.join([unicode(x) for x in VERSION]))
    do_main(arguments)


def print_err(s):
    sys.stderr.write(s + '\n')


def cprint(string, color='WHITE', verbose=False):
    if sys.stdout.isatty():
        term = terminal.get_terminal()
        term.set_color(terminal.colors[color])
        if verbose:
            if SHOW_VERBOSE:
                sys.stdout.write(string)
        else:
            sys.stdout.write(string)
        term.reset()
    else:
        if not verbose:
            sys.stdout.write(string)


if __name__ == '__main__':
    main()
