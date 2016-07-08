#!/usr/bin/env python

import json
import os
import sys
import uuid
from copy import deepcopy
from xml.etree import ElementTree as ET

import requests
from requests.packages.urllib3.exceptions import InsecureRequestWarning

SCHEMA = "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2"
XMLNS_S = "{http://www.w3.org/2003/05/soap-envelope}"
XMLNS_WSEN = "{http://schemas.xmlsoap.org/ws/2004/09/enumeration}"
XMLNS_WSMAN = "{http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd}"
XMLNS_N1_BASE = "{http://schemas.dell.com/wbem/wscim/1/cim-schema/2/%s}"
# Generic wsman-crafted soap message
SOAP_ENUM_WSMAN_TEMPLATE = '''<?xml version="1.0"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:wsman="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:wsen="http://schemas.xmlsoap.org/ws/2004/09/enumeration">
  <s:Header>
    <wsa:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate</wsa:Action>
    <wsa:To s:mustUnderstand="true">{management_url}</wsa:To>
    <wsman:ResourceURI s:mustUnderstand="true">{resource}</wsman:ResourceURI>
    <wsa:MessageID s:mustUnderstand="true">uuid:{uuid}</wsa:MessageID>
    <wsa:ReplyTo>
      <wsa:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:Address>
    </wsa:ReplyTo>
    <wsman:SelectorSet>
      <wsman:Selector Name="__cimnamespace">{selector}</wsman:Selector>
    </wsman:SelectorSet>
  </s:Header>
  <s:Body>
    <wsen:Enumerate>
      <wsman:OptimizeEnumeration/>
      <wsman:MaxElements>{max_elements}</wsman:MaxElements>
    </wsen:Enumerate>
  </s:Body>
</s:Envelope>
'''

MAC_PREFIX_BLACKLIST = [
    '505054', '33506F', '009876', '000000', '00000C', '204153', '149120',
    '020054', 'FEFFFF', '1AF920', '020820', 'DEAD2C', 'FEAD4D',
]
SERIAL_BLACKLIST = [
    None, '', 'Not Available', 'XxXxXxX', '-----', '[Unknown]', '0000000000',
    'Not Specified', 'YK10CD', '1234567890', 'None', 'To Be Filled By O.E.M.',
]

DEVICE_INFO_TEMPLATE = {
    "model_name": "",
    "ethernets": [],
    "memory": [],
    "fibre_channel_cards": [],
    "processors": [],
    "disks": [],
    "serial_number": "",
}
ETHERNET_TEMPLATE = {
    "mac": "",
    "model_name": "",
    "speed": "unknown speed",
    "firmware_version": "",
}
PROCESSOR_TEMPLATE = {
    "model_name": "",
    "family": "",
    "label": "",
    "index": None,
    "speed": None,
    "cores": None,
}
MEMORY_TEMPLATE = {
    "model_name": "",
    "size": None,
    "speed": None,
}
FIBRE_CHANNEL_CARD_TEMPLATE = {
    # firmware_version, speed and wwn are unused (iDRAC doesn't provide this info yet).
    "firmware_version": "",
    "model_name": "",
    "speed": "unknown speed",
    "wwn": "",
}


def normalize_mac_address(mac_address):
    mac_address = mac_address.upper().replace('-', ':')
    return mac_address


requests.packages.urllib3.disable_warnings(InsecureRequestWarning)


class IdracError(Exception):
    pass


class IDRAC(object):

    def __init__(self, host, user, password):
        self.host = host
        self.user = user
        self.password = password

    def run_command(self, class_name, selector='root/dcim'):
        management_url = "https://{}/wsman".format(self.host)
        generated_uuid = uuid.uuid1()
        message = SOAP_ENUM_WSMAN_TEMPLATE.format(
            resource=SCHEMA.rstrip('/') + '/' + class_name,
            management_url=management_url,
            uuid=generated_uuid,
            selector=selector,
            max_elements=255,
        )
        return ET.XML(self._send_soap(
            url='https://{}/wsman'.format(self.host),
            message=message,
        ))

    def _send_soap(self, url, message):
        """Try to send soap message to post_url using http basic
        authentication. Note, that we don't store any session information,
        nor validate SSL certificate. Any following requests will re-send
        basic auth header again.
        """
        r = requests.post(
            url,
            data=message,
            auth=(self.user, self.password),
            verify=False,
            headers={
                'Content-Type': 'application/soap+xml;charset=UTF-8',
            },
        )
        if not r.ok:
            if r.status_code == 401:
                raise IdracError("Auth error. Invalid username or password.")
            raise IdracError(
                "SoapError: Reponse: {}\nRequest: {}".format(
                    r.text, message,
                ),
            )
        errors_path = '{s}Body/{s}Fault'.format(s=XMLNS_S)
        errors_list = []
        errors_node = ET.XML(r.text).find(errors_path)
        if errors_node:
            errors_list = [node_text for node_text in errors_node.itertext()]
            raise IdracError(
                'SoapError: Request: {}, Response errors: {}'.format(
                    message, ','.join(errors_list),
                )
            )
        return r.text


def _get_base_info(idrac_manager):
    device_info = DEVICE_INFO_TEMPLATE
    tree = idrac_manager.run_command('DCIM_SystemView')
    xmlns_n1 = XMLNS_N1_BASE % "DCIM_SystemView"
    q = "{}Body/{}EnumerateResponse/{}Items/{}DCIM_SystemView".format(
        XMLNS_S,
        XMLNS_WSEN,
        XMLNS_WSMAN,
        xmlns_n1,
    )
    records = tree.findall(q)
    if not records:
        raise IdracError("Incorrect answer in the _get_base_info.")
    model_name = "{} {}".format(
        records[0].find(
            "{}{}".format(xmlns_n1, 'Manufacturer'),
        ).text.strip().replace(" Inc.", ""),
        records[0].find(
            "{}{}".format(xmlns_n1, 'Model'),
        ).text.strip(),
    )
    serial_number = records[0].find(
        "{}{}".format(xmlns_n1, 'ChassisServiceTag'),
    ).text.strip()
    if serial_number not in SERIAL_BLACKLIST:
        device_info['serial_number'] = serial_number
    device_info['model_name'] = model_name
    return device_info


def _get_ethernets(idrac_manager):

    def get_mac():
        # e.g. CurrentMACAddress: BB:CC:DD:EE:FF:00
        try:
            mac = record.find(
                "{}{}".format(xmlns_n1, 'CurrentMACAddress'),
            ).text
        except AttributeError:
            mac = None
        if mac:
            mac = normalize_mac_address(mac)
            if mac[:6] in MAC_PREFIX_BLACKLIST:
                mac = None
        return mac

    def get_model(mac):
        # e.g. ProductName: Intel(R) Ethernet 10G 4P X520/I350 rNDC - BB:CC:DD:EE:FF:00
        try:
            model = record.find(
                "{}{}".format(xmlns_n1, 'ProductName'),
            ).text
        except AttributeError:
            model = None
        if mac in model:
            mm = []
            for m in model.split():
                if len(m) == 1 or mac in m:
                    continue
                mm.append(m)
            model = ' '.join(mm)
        return model

    def get_speed(model):
        # Unfortunately, it seems that there's no separate field for this, so
        # we have to take it from ProductName/model.
        if not model:
            speed = "unknown speed"
        elif "Gigabit" in model:
            speed = "1 Gbps"
        else:
            speeds = {
                '10M':  "10 Mbps",
                '100M': "100 Mbps",
                '1G':   "1 Gbps",
                '10G':  "10 Gbps",
                '40G':  "40 Gbps",
                '100G': "100 Gbps",
            }
            for s in speeds.keys():
                if s in model:
                    speed = speeds[s]
                    break
            else:
                speed = "unknown speed"
        return speed

    def get_fw_version():
        # e.g. FamilyVersion: 15.0.27
        try:
            ver = record.find(
                "{}{}".format(xmlns_n1, 'FamilyVersion'),
            ).text
        except AttributeError:
            ver = None
        return ver

    tree = idrac_manager.run_command('DCIM_NICView')
    xmlns_n1 = XMLNS_N1_BASE % "DCIM_NICView"
    q = "{}Body/{}EnumerateResponse/{}Items/{}DCIM_NICView".format(
        XMLNS_S,
        XMLNS_WSEN,
        XMLNS_WSMAN,
        xmlns_n1,
    )
    ethernets = []
    for record in tree.findall(q):
        mac = get_mac()
        if not mac:
            continue
        model = get_model(mac)
        speed = get_speed(model)
        ver = get_fw_version()
        ethernet = {
            "mac": mac,
            "model_name": model,
            "speed": speed,
            "firmware_version": ver,
        }
        ethernets.append(ethernet)
    return ethernets


def _get_processors(idrac_manager):
    tree = idrac_manager.run_command('DCIM_CPUView')
    xmlns_n1 = XMLNS_N1_BASE % "DCIM_CPUView"
    q = "{}Body/{}EnumerateResponse/{}Items/{}DCIM_CPUView".format(
        XMLNS_S,
        XMLNS_WSEN,
        XMLNS_WSMAN,
        xmlns_n1,
    )
    results = []
    for record in tree.findall(q):
        model = record.find("{}{}".format(xmlns_n1, 'Model')).text.strip()
        try:
            index = int(
                record.find(
                    "{}{}".format(xmlns_n1, 'InstanceID'),
                ).text.strip().split('.')[-1],
            )
        except (ValueError, IndexError):
            continue
        results.append({
            'cores': int(record.find(
                "{}{}".format(xmlns_n1, 'NumberOfProcessorCores'),
            ).text.strip()),
            'model_name': model,
            'speed': int(record.find(
                "{}{}".format(xmlns_n1, 'MaxClockSpeed'),
            ).text.strip()),
            'index': index,
            'family': record.find(
                "{}{}".format(xmlns_n1, 'CPUFamily'),
            ).text.strip(),
            'label': model,
        })
    return results


def _get_memory(idrac_manager):
    tree = idrac_manager.run_command('DCIM_MemoryView')
    xmlns_n1 = XMLNS_N1_BASE % "DCIM_MemoryView"
    q = "{}Body/{}EnumerateResponse/{}Items/{}DCIM_MemoryView".format(
        XMLNS_S,
        XMLNS_WSEN,
        XMLNS_WSMAN,
        xmlns_n1,
    )
    return [
        {
            'model_name': '{} {}'.format(
                record.find(
                    "{}{}".format(xmlns_n1, 'Manufacturer'),
                ).text.strip(),
                record.find(
                    "{}{}".format(xmlns_n1, 'Model'),
                ).text.strip(),
            ),
            'size': int(record.find(
                "{}{}".format(xmlns_n1, 'Size'),
            ).text.strip()),
            'speed': int(record.find(
                "{}{}".format(xmlns_n1, 'Speed'),
            ).text.strip()),
        } for record in tree.findall(q)
    ]


def _get_disks(idrac_manager):
    tree = idrac_manager.run_command('DCIM_PhysicalDiskView')
    xmlns_n1 = XMLNS_N1_BASE % "DCIM_PhysicalDiskView"
    q = "{}Body/{}EnumerateResponse/{}Items/{}DCIM_PhysicalDiskView".format(
        XMLNS_S,
        XMLNS_WSEN,
        XMLNS_WSMAN,
        xmlns_n1,
    )
    results = []
    for record in tree.findall(q):
        manufacturer = record.find(
            "{}{}".format(xmlns_n1, 'Manufacturer'),
        ).text.strip()
        model_name = '{} {}'.format(
            manufacturer,
            record.find(
                "{}{}".format(xmlns_n1, 'Model')
            ).text.strip(),
        )
        size_in_bytes = record.find(
            '{}{}'.format(xmlns_n1, 'SizeInBytes'),
        ).text
        serial_number = record.find(
            '{}{}'.format(xmlns_n1, 'SerialNumber'),
        ).text
        results.append({
            'size': int(
                int(
                    size_in_bytes and size_in_bytes.strip() or 0
                ) / 1024 / 1024 / 1024
            ),
            'serial_number': serial_number and serial_number.strip() or '',
            'label': model_name,
            'model_name': model_name,
            'family': manufacturer,
        })
    return results


def _get_fibre_channel_cards(idrac_manager):
    tree = idrac_manager.run_command('DCIM_PCIDeviceView')
    xmlns_n1 = XMLNS_N1_BASE % "DCIM_PCIDeviceView"
    q = "{}Body/{}EnumerateResponse/{}Items/{}DCIM_PCIDeviceView".format(
        XMLNS_S,
        XMLNS_WSEN,
        XMLNS_WSMAN,
        xmlns_n1,
    )
    fc_cards = []
    for record in tree.findall(q):
        model_name = record.find(
            "{}{}".format(xmlns_n1, "Description"),
        ).text
        if 'fibre channel' not in model_name.lower():
            continue
        fc_card = deepcopy(FIBRE_CHANNEL_CARD_TEMPLATE)
        fc_card['model_name'] = model_name
        fc_cards.append(fc_card)
    return fc_cards


def idrac_device_info(idrac_manager):
    device_info = _get_base_info(idrac_manager)
    device_info['ethernets'] = _get_ethernets(idrac_manager)
    device_info['processors'] = _get_processors(idrac_manager)
    device_info['memory'] = _get_memory(idrac_manager)
    device_info['disks'] = _get_disks(idrac_manager)
    device_info['fibre_channel_cards'] = _get_fibre_channel_cards(idrac_manager)
    return device_info


def scan(host, user, password):
    if host == "":
        raise IdracError("No IP address to scan has been provided.")
    if user == "":
        raise IdracError("No management username has been provided.")
    if password == "":
        raise IdracError("No management password has been provided.")
    idrac_manager = IDRAC(host, user, password)
    device_info = idrac_device_info(idrac_manager)
    print(json.dumps(device_info))


if __name__ == '__main__':
    host = os.environ.get('IP_TO_SCAN', "")
    user = os.environ.get('MANAGEMENT_USER_NAME', "")
    password = os.environ.get('MANAGEMENT_USER_PASSWORD', "")
    try:
        scan(host, user, password)
    except IdracError as e:
        print(e.args[0])
        sys.exit(1)
