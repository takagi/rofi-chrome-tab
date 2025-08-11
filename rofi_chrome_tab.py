#!/usr/bin/python

import asyncio
import json
import os
import sys
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Callable, cast

#
# Util
#

def log(msg: str) -> None:
    with open('/tmp/rofi-chrome-tab.log', 'a') as f:
        now = datetime.now(timezone.utc)
        print(f'[{now.isoformat()}] {msg}', file=f)

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
    if message_length == 0:
        raise EOFError()

    encoded_message: bytes = sys.stdin.buffer.read(message_length)
    message: str = encoded_message.decode('utf-8')
    log('recv_message: ' + message[:30] + ('...' if len(message) >= 30 else ''))

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

tabs: list[Tab] = []

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
# UDS handler
#

async def handle_client(reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
    log('handle_client()')

    input = (await reader.read(1024)).decode('utf-8').strip()
    log('recv: ' + input)

    try:
        command = parse_command(input)
    except ValueError as e:
        writer.write(b'Invalid command\n')
        raise e

    if isinstance(command, CountCommand):
        count = NativeMessaging.count_tabs()
        writer.write(str(count).encode())
        writer.write(b'\n')
        await writer.drain()

    elif isinstance(command, ListCommand):
        for tab in tabs:
            line = ','.join([str(PID), str(tab_id(tab)), tab_host(tab), tab_title(tab)])
            writer.write(line.encode())
            writer.write(b'\n')
            await writer.drain()

    elif isinstance(command, SelectCommand):
        NativeMessaging.select_tab(command.tabId)

    else:
        raise AssertionError

    writer.close()
    await writer.wait_closed()

#
# stdin callback
#

def stdin_callback() -> None:
    log('stdin_callback()')

    try:
        _ = recv_message()  # updated
    except EOFError:
        # Shutting down
        loop = asyncio.get_running_loop()
        loop.remove_reader(sys.stdin)
        return

    time.sleep(0.5)  # delay before a tab is switched

    global tabs
    tabs = NativeMessaging.list_tabs()

#
# Main
#

PID = os.getpid()
SOCKET_PATH = f'/tmp/native-app.{PID}.sock'

if os.path.exists(SOCKET_PATH):
    os.remove(SOCKET_PATH)

async def main() -> None:
    server = await asyncio.start_unix_server(handle_client, SOCKET_PATH)

    loop = asyncio.get_running_loop()
    loop.add_reader(sys.stdin, stdin_callback)

    async with server:
        await server.serve_forever()

if __name__ == '__main__':
    asyncio.run(main())
