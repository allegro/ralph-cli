#!/usr/bin/env python

import json
import os
import sys
from copy import deepcopy

import hpilo


MAC_PREFIX_BLACKLIST = [
    '505054', '33506F', '009876', '000000', '00000C', '204153', '149120',
    '020054', 'FEFFFF', '1AF920', '020820', 'DEAD2C', 'FEAD4D',
]

DEVICE_INFO_TEMPLATE = {
    "model_name": "",
    "processors": [],
    "ethernets": [],
    "disks": [],          # unused (hpilo doesn't provide such info)
    "serial_number": "",
    "memory": [],
}

ETHERNET_TEMPLATE = {
    # model_name, speed and firmware are unused (hpilo doesn't provide such info)
    "mac": "",
    "model_name": "",
    "speed": "unknown speed",
    "firmware_version": "",

}

PROCESSOR_TEMPLATE = {
    "model_name": "",  # unused (hpilo doesn't provide such info)
    "family": "",      # hpilo returns int here, but only for iLO2 (for iLO3 this field is empty)  # noqa
    "label": "",
    "index": None,     # unused, but similar info is available as "label"
    "speed": None,
    "cores": None,
}

MEMORY_TEMPLATE = {
    "model_name": "",  # unused (hpilo doesn't provide such info)
    "size": None,
    "speed": None,
}


class IloError(Exception):
    pass


def normalize_mac(mac):
    mac = mac.upper().replace('-', ':')
    return mac


def get_ilo_instance(host, user, password):
    ilo = hpilo.Ilo(hostname=host, login=user, password=password)
    return ilo


def _get_ethernets(raw_macs, ilo_version):
    # The data structure for MAC addresses returned from hpilo is pretty nasty,
    # especially for iLO3 (no clear distinction between embedded NICs and
    # iSCSI ports).
    if ilo_version == 3:
        start_idx = 0
    else:
        start_idx = 1
    ethernets = []
    for m in raw_macs:
        fields = m.get('fields', [])
        for i in range(start_idx, len(fields), 2):
            if (
                fields[i]['name'] == 'Port' and
                fields[i]['value'] != 'iLO'  # belongs to mgmt address
            ):
                mac = normalize_mac(fields[i + 1]['value'])
                if mac[:6] not in MAC_PREFIX_BLACKLIST:
                    eth = deepcopy(ETHERNET_TEMPLATE)
                    eth["mac"] = mac
                    ethernets.append(eth)
    return ethernets


def _get_speed(s):
    # sample return value from hpilo: "2533 MHz"
    if s is not None:
        s = int(s.split(" ")[0])
    return s


def _get_processors(raw_procs):

    def get_cores(c):
        # sample return value from hpilo: "4 of 4 cores; 8 threads"
        if c is not None:
            c = int(c.split(" of ")[0])
        return c

    processors = []
    for p in raw_procs:
        proc = deepcopy(PROCESSOR_TEMPLATE)
        proc['family'] = str(p.get('Family', ""))
        proc['label'] = p.get('Label', "")
        proc['speed'] = _get_speed(p.get('Speed'))
        proc['cores'] = get_cores(p.get('Execution Technology'))
        processors.append(proc)
    return processors


def _get_memory(raw_memory):

    def get_size(s):
        # sample return value from hpilo: "4096 MB"
        if s is not None:
            s = int(s.split(" ")[0])
        return s

    memory = []
    for m in raw_memory:
        mem = deepcopy(MEMORY_TEMPLATE)
        mem['size'] = get_size(m.get('Size'))
        mem['speed'] = _get_speed(m.get('Speed'))
        memory.append(mem)
    return memory


# The data structure returned from python-hpilo is quite inconvenient for our
# use-case, therefore we need to reshape it a little bit.
def _prepare_host_data(raw_host_data, ilo_version):
    host_data = {
        "sys_info": [],
        "processors": [],
        "memory": [],
        "mac_addresses": [],
    }
    if ilo_version == 2:
        for part in raw_host_data:
            if part.get('Subject') == 'System Information':
                host_data['sys_info'].append(part)
                continue
            if part.get('Subject') == 'Processor Information':
                host_data['processors'].append(part)
                continue
            if (
                    part.get('Subject') == 'Memory Device' and
                    part.get('Size') != 'not installed'
            ):
                host_data['memory'].append(part)
                continue
            if part.get('Subject') is None and part.get('fields') is not None:
                fields = part.get('fields')
                for field in fields:
                    if (
                            isinstance(field, dict) and
                            field['value'] == 'Embedded NIC MAC Assignment'
                    ):
                        host_data['mac_addresses'].append(part)
                        break
                continue
    elif ilo_version == 3:
        for part in raw_host_data:
            if part.get('Product Name') is not None:
                host_data['sys_info'].append(part)
                continue
            if part.get('Execution Technology') is not None:
                host_data['processors'].append(part)
                continue
            if (
                part.get('Label') is not None and
                part.get('Size') is not None and
                part.get('Speed') is not None
            ):
                host_data['memory'].append(part)
                continue
            if part.get('fields') is not None:
                for field in part.get('fields'):
                    # The condition here is not very reliable, but that's the
                    # only way to distinguish between 'fields' list containing
                    # embedded NICs vs. iSCSIs ports.
                    if (
                            isinstance(field, dict) and
                            field['name'] == 'Port' and
                            field['value'] == 'iLO'
                    ):
                        host_data['mac_addresses'].append(part)
                        break
                continue
    else:
        raise IloError("Unknown version of iLO: %d".format(ilo_version))
    if len(host_data['sys_info']) > 1:
        raise IloError(
            "There should be only one 'System Information' dict "
            "in the data returned by python-hpilo."
        )
    return host_data


def get_ilo_version(ilo_manager):
    fw_version = ilo_manager.get_fw_version()
    if fw_version.get('management_processor') == "iLO3":
        ilo_version = 3
    elif fw_version.get('management_processor') == "iLO2":
        ilo_version = 2
    else:
        ilo_version = None
    return ilo_version


def ilo_device_info(ilo_manager, ilo_version):
    raw_host_data = ilo_manager.get_host_data()
    host_data = _prepare_host_data(raw_host_data, ilo_version)
    device_info = DEVICE_INFO_TEMPLATE
    device_info['processors'] = _get_processors(host_data['processors'])
    device_info['ethernets'] = (
        _get_ethernets(host_data['mac_addresses'], ilo_version)
    )
    device_info['serial_number'] = (
        host_data['sys_info'][0].get('Serial Number', "").strip()
    )
    device_info['model_name'] = (
        host_data['sys_info'][0].get('Product Name', "")
    )
    device_info['memory'] = _get_memory(host_data['memory'])
    return device_info


def scan(host, user, password):
    if host == "":
        raise IloError("No IP address to scan has been provided.")
    if user == "":
        raise IloError("No management username has been provided.")
    if password == "":
        raise IloError("No management password has been provided.")
    ilo_manager = get_ilo_instance(host, user, password)
    ilo_version = get_ilo_version(ilo_manager)
    device_info = ilo_device_info(ilo_manager, ilo_version)
    print(json.dumps(device_info))


if __name__ == '__main__':
    host = os.environ.get('IP_TO_SCAN', "")
    user = os.environ.get('MANAGEMENT_USER_NAME', "")
    password = os.environ.get('MANAGEMENT_USER_PASSWORD', "")
    try:
        scan(host, user, password)
    except (IloError, hpilo.IloCommunicationError) as e:
        print(e.args[0])
        sys.exit(1)
