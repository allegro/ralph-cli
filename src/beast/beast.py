#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""Beast is convenient Ralph API commandline client.
Documentation on https://github.com/allegro/ralph_beast/
Example usage $ ~/beast show

Usage:
 beast show
 beast show <resource> [--schema] [--fields=fields] [--filter-fields] [--filter=field_filter] [--limit=limit] [--csv] [--trim] [--width=max_width]
 beast update <resource> <id> <fields> <fields_values>
 beast -h | --help
 beast --version

"""

import csv
import os
import requests
import slumber
import sys
import time
import urlparse

from collections import defaultdict
from docopt import docopt


class Api(object):
    def get_session(self, settings):
        url = settings.get('url')
        version = settings.get('version')
        username = settings.get('username')
        api_key = settings.get('api_key')
        if not username or not api_key or not url:
            print("username, api_key, url and version in ~/.beast/config are required.")
            sys.exit(2)
        session = requests.session(verify=False)
        session = slumber.API(
            '%(url)s/api/%(version)s/' % dict(url=url, version=version),
            session=session,
        )
        return session

    def put_resource(self, settings, resource, id, data):
        session = self.get_session(settings)
        username = settings.get('username')
        api_key = settings.get('api_key')
        try:
            data = getattr(session, resource)(id).put(
                data=data,
                username=username,
                api_key=api_key
            )
        except slumber.exceptions.HttpClientError as error:
            print 'Error: ', error.content
        return data

    def get_resource(self, settings, resource, limit=None, filters=None):
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
        return getattr(session, resource).get(**dict(attrs_dict))

    def get_schema(self, settings, resource=None, filters=False):
        if not resource:
            print "-" * 50
            data = self.get_resource(settings, '')
            list_of_resources = [resource for resource in data]
            print '\n'.join(sorted(list_of_resources))
        else:
            print "-" * 50
            schema = '%s/schema/' % resource
            data = self.get_resource(settings, resource=schema, limit=1)
            if not filters:
                for field in data['fields']:
                    print field
            else:
                for field in data['filtering'].keys():
                    print field


class Content(object):
    def get_api_objects(self, data, output_fields):
        if not 'objects' in data:
            return data, []
        content = []
        for row in data['objects']:
            header = output_fields if output_fields else row.keys()
            content.append(self.remove_links(row))
        return header, content

    def console_repr(self, field):
        if type(field) == type(u''):
            if '/api/v0.9/' in field:
                field = unicode(field.replace('/api/v0.9/',''))
            else:
                field = unicode(field[:10])
        elif type(field) == type(False):
            field = 'x' if field else 'o'
        elif type(field) == type({}):
            field = field['id']
        elif type(field) == type(None):
            field = ''
        elif type(field) == type([]):
            field = ','.join([self.console_repr(subfield) for subfield in field])
        else:
            field = unicode(field)
        return field

    def remove_links(self, row):
        for field in row.keys():
            encode_field = field.encode('utf-8')
            row[encode_field] = self.console_repr(row[encode_field])
        return row

    def smallest_list_of(self, widths, output_fields, max_width):
        of_truncated = []
        current_width = 0
        for key in output_fields:
            width = widths.get(key)
            if current_width + width > max_width:
                return of_truncated
            else:
                of_truncated.append(key)
            current_width += width
        return of_truncated

    def trim_colums(self, of, widths, trim_columns, result_data,
                                                    output_fields, max_width):
        # TODO: FIX ME!
        for row in of:
            widths[row] = len(row) + 4 if not trim_columns else 5

        if not output_fields:
            of = self.smallest_list_of(widths, of, max_width)
        else:
            for row in result_data:
                for key in row.keys():
                    widths[key] = max(widths[key], len(row[key])+4)
                    all_fields = row.keys()
                    of = output_fields if output_fields else all_fields
        return of, widths


class Writer(Content):
    def write_header(self, of, result_data, trim_columns, output_fields,
                                                            max_width, widths):
        of, widths = self.trim_colums(
            of,
            widths,
            trim_columns,
            result_data,
            output_fields,
            max_width
        )
        print "-" * max_width
        sys.stdout.write("|")
        for key in of:
            fill = widths.get(key)
            align = widths.get(key)
            sys.stdout.write(
                ' {: <{fill}.{align}}|'.format(
                    key,
                    fill=fill-4,
                    align=align-4
                )
            )
        sys.stdout.write('\n')

    def write_rows(self, of, result_data, trim_columns, output_fields,
                                                            max_width, widths):
        of, widths = self.trim_colums(
            of,
            widths,
            trim_columns,
            result_data,
            output_fields,
            max_width
        )
        for row in result_data:
            print "-" * max_width
            sys.stdout.write("|")
            for key in of:
                fill = widths.get(key)
                align = widths.get(key)
                sys.stdout.write(
                    ' {: <{fill}.{align}}|'.format(
                        row[key].encode('utf-8'),
                        fill=fill-4,
                        align=align-4
                    )
                )
            sys.stdout.write('\n')

    def csv(self, header, content):
        print ';'.join(field for field in header)
        for fields in content:
            print ';'.join(
                field.encode('utf-8') for key, field in fields.items()
            )


def show(arguments, settings):
    resource = arguments.get('<resource>', '')
    limit = arguments.get('--limit')
    python_filter = arguments.get('--python_filter')
    fields = arguments.get('--fields')
    out_fls = [field.strip() for field in fields.split(',')] if fields else None
    max_width = int(arguments.get('--width') or 120)

    if not resource:
        print "Ralph API, schema"
        return Api().get_schema(settings, None)
    elif arguments.get('--schema'):
        print "Ralph API > %s, schema" % resource
        return Api().get_schema(settings, resource)

    if arguments.get('--filter-fields'):
        print "Ralph API > %s, available filter fileds" % resource
        return Api().get_schema(settings, resource, filters=True)

    data = Api().get_resource(
        settings,
        resource,
        limit,
        arguments.get('--filter'),
    )
    if python_filter:
        print "Python filter: %s" % python_filter
        header, content = Content().python_filter(data, python_filter)
    else:
        print "Ralph API > %s" % resource
        header, content = Content().get_api_objects(data, out_fls)

    if limit:
        print "Limit: %s" % limit

    if arguments.get('--csv'):
        return Writer().csv(header, content)

    Writer().write_header(
        header,
        content,
        arguments.get('--trim'),
        out_fls,
        max_width,
        defaultdict(int),
    )
    Writer().write_rows(
        header,
        content,
        arguments.get('--trim'),
        out_fls,
        max_width,
        defaultdict(int),
    )


def update(arguments, settings):
    resource = arguments.get('<resource>')
    id = arguments.get('<id>')
    fields = arguments.get('<fields>').split(',')
    fields_values = arguments.get('<fields_values>')
    for row in csv.reader([fields_values],
        quotechar='"', quoting=csv.QUOTE_MINIMAL):
        pass
    data = dict(zip(fields, row))
    Api().put_resource(settings, resource, id, data)


def do_main(arguments):
    settings = dict()
    config_file = os.path.expanduser("~/.beast/config")
    try:
        execfile(config_file, settings)
    except IOError:
        print("Config file ~/.beast/config doesn't exist.")
        sys.exit(4)
    if arguments.get('show'):
        show(arguments, settings)
    elif arguments.get('update'):
        update(arguments, settings)


def main():
    stopwatch_start = time.time()
    arguments = docopt(__doc__, version='0.1')
    do_main(arguments)
    print '\nRuntime: %s sec' % round(time.time()-stopwatch_start, 2)


if __name__ == '__main__':
    main()
