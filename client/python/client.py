import json
import http.client
import urllib.parse

from typing import List
from dataclasses import dataclass

class AddRequest:
    def __init__(self, host_prefix: str, target: str):
        self.host_prefix = host_prefix
        self.target = target

    def to_json(self) -> str:
        return json.dumps(self.__dict__)

    @classmethod
    def from_json(cls, json_str: str) -> 'AddRequest':
        data = json.loads(json_str)
        return cls(**data)
    
class AddResponse:
    def __init__(self, hostname: str):
        self.hostname = hostname

    def to_json(self) -> str:
        return json.dumps(self.__dict__)

    @classmethod
    def from_json(cls, json_str: str) -> 'AddResponse':
        data = json.loads(json_str)
        return cls(**data)
    
class ListHost:
    def __init__(self, hostname: str, target: str):
        self.hostname = hostname
        self.target = target

    def to_json(self) -> str:
        return json.dumps(self.__dict__)

    @classmethod
    def from_json(cls, json_str: str) -> 'ListHost':
        data = json.loads(json_str)
        return cls(**data)

class ListResponse:
    def __init__(self, hosts: List[ListHost]):
        self.hosts = hosts

    def to_json(self) -> str:
        return json.dumps(self.__dict__)

    @classmethod
    def from_json(cls, json_str: str) -> 'ListResponse':
        data = json.loads(json_str)
        return cls(**data)


@dataclass
class Client:
    base: str
    cl: http.client.HTTPConnection

    @classmethod
    def from_url(cls, url: str) -> 'Client':
        parsed_url = urllib.parse.urlparse(url)
        conn = http.client.HTTPConnection(parsed_url.netloc)
        return cls(parsed_url.path, conn)

    def add(self, host_prefix: str, target: str) -> str:
        req_body = json.dumps(AddRequest(host_prefix=host_prefix, target=target)).encode('utf-8')
        headers = {'Content-Type': 'application/json'}
        self.cl.request('POST', self.base + '/vhost/', body=req_body, headers=headers)
        resp = self.cl.getresponse()
        if resp.status != http.client.OK:
            raise Exception(f'unexpected status code: {resp.status}')
        resp_body = resp.read().decode('utf-8')
        add_resp = AddResponse.from_json(resp_body)
        return add_resp.hostname

    def remove(self, hostname: str) -> None:
        self.cl.request('DELETE', self.base + f'/vhost/{hostname}')
        resp = self.cl.getresponse()
        if resp.status != http.client.OK:
            raise Exception(f'unexpected status code: {resp.status}')

    def list(self) -> List[ListHost]:
        self.cl.request('GET', self.base + '/vhost/')
        resp = self.cl.getresponse()
        if resp.status != http.client.OK:
            raise Exception(f'unexpected status code: {resp.status}')
        resp_body = resp.read().decode('utf-8')
        list_resp = ListResponse.from_json(resp_body)
        return list_resp.hosts
