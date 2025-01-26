#!/usr/bin/python

from dataclasses import dataclass
import json
import os
import socketserver
import sys
import time
from typing import cast, Callable

#
# Util
#

def log(msg: str) -> None:
    print(msg, file=sys.stderr)

def call_with_retry[T](thunk: Callable[[], T]) -> T:
    for _ in range(3):
        try:
            return thunk()
        except Exception as e:
            err = e
            log(str(e))
            log('retrying...')
            time.sleep(3)
    else:
        raise err

#
# Native messaging
#

def recv_message() -> object:
    raw_length: bytes = sys.stdin.buffer.read(4)
    message_length = int.from_bytes(raw_length, 'little')
    encoded_message: bytes = sys.stdin.buffer.read(message_length)

    message: str = encoded_message.decode('utf-8')
    log('recv_message: ' + message)
    return json.loads(message)

def send_message(message: object) -> None:
    json_message: str = json.dumps(message)
    log('send_message: ' + json_message)

    encoded_message: bytes = json_message.encode()
    sys.stdout.buffer.write(len(encoded_message).to_bytes(4, 'little'))
    sys.stdout.buffer.write(encoded_message)
    sys.stdout.flush()

#
# Tab
#

Tab = dict[str, int | str]

def tab_id(tab: Tab) -> int:
    return cast(int, tab['id'])

def tab_title(tab: Tab) -> str:
    return cast(str, tab['title'])

def tab_host(tab: Tab) -> str:
    return cast(str, tab['host'])

#
# NativeMessaging
#

class NativeMessaging:

    @staticmethod
    def count_tabs() -> int:
        send_message({'command': 'count'})
        return cast(int, recv_message())

    @staticmethod
    def list_tabs() -> list[Tab]:
        send_message({'command': 'list'})
        return cast(list[Tab], recv_message())

    @staticmethod
    def select_tab(id: int) -> None:
        send_message({'command': 'select', 'tabId': id})
        recv_message()  # ok

#
# Command
#

@dataclass
class BaseCommand:
    pass

@dataclass
class CountCommand(BaseCommand):
    pass

@dataclass
class ListCommand(BaseCommand):
    pass

@dataclass
class SelectCommand(BaseCommand):
    tabId: int

def parse_command(input: str) -> BaseCommand:
    if input == 'count':
        return CountCommand()

    if input == 'list':
        return ListCommand()

    if input.startswith('select '):
        tabId = int(input.split()[1])
        return SelectCommand(tabId)

    raise ValueError(f'Invalid command: {input}')

#
# Server
#

PID = os.getpid()
SOCKET_PATH = f'/tmp/native-app.{PID}.sock'

if os.path.exists(SOCKET_PATH):
    os.remove(SOCKET_PATH)

class RequestHandler(socketserver.BaseRequestHandler):
    def handle(self) -> None:
        input = self.request.recv(1024).strip().decode('utf-8')
        log('recv: ' + input)

        try:
            command = parse_command(input)
        except ValueError as e:
            self.request.send(b'Invalid command\n')
            raise e

        if isinstance(command, CountCommand):
            count = NativeMessaging.count_tabs()
            self.request.send(str(count).encode())
            self.request.send(b'\n')

        elif isinstance(command, ListCommand):
            tabs = NativeMessaging.list_tabs()

            for tab in tabs:
                line = ','.join([str(PID), str(tab_id(tab)), tab_host(tab), tab_title(tab)])
                self.request.send(line.encode())
                self.request.send(b'\n')

        elif isinstance(command, SelectCommand):
            NativeMessaging.select_tab(command.tabId)

        else:
            raise AssertionError

#
# Main
#

def main() -> None:
    def make_server() -> socketserver.UnixStreamServer:
        return socketserver.UnixStreamServer(SOCKET_PATH, RequestHandler)
    with call_with_retry(make_server) as server:
        server.serve_forever()

if __name__ == '__main__':
    main()
