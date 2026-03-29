#!/usr/bin/env python3

import asyncio
import json
import sys
import websockets
from typing import Optional


class ChatClient:

    def __init__(self, server_url: str = "ws://localhost:8080/ws"):
        self.server_url = server_url
        self.websocket: Optional[websockets.WebSocketClientProtocol] = None
        self.nickname: str = "Anonymous"
        self.connected: bool = False

    async def connect(self, nickname: str = "Anonymous") -> bool:
        try:
            self.websocket = await websockets.connect(self.server_url)
            self.nickname = nickname
            self.connected = True

            message = {
                "type": "nick_change",
                "new_nickname": nickname
            }
            await self.websocket.send(json.dumps(message))

            print(f"[System] Connected to server as '{nickname}'")
            return True
        except Exception as e:
            print(f"[Error] Failed to connect: {e}")
            return False

    async def disconnect(self):
        if self.websocket:
            await self.websocket.close()
            self.connected = False
            print("[System] Disconnected from server")

    async def send_message(self, content: str) -> bool:
        if not self.connected:
            print("[Error] Not connected to server")
            return False

        message = {
            "type": "chat",
            "content": content
        }
        try:
            await self.websocket.send(json.dumps(message))
            return True
        except Exception as e:
            print(f"[Error] Failed to send message: {e}")
            return False

    async def change_nickname(self, new_nickname: str) -> bool:
        if not self.connected:
            print("[Error] Not connected to server")
            return False

        message = {
            "type": "nick_change",
            "new_nickname": new_nickname
        }
        try:
            await self.websocket.send(json.dumps(message))
            return True
        except Exception as e:
            print(f"[Error] Failed to change nickname: {e}")
            return False

    async def receive_messages(self):
        if not self.connected:
            return

        try:
            async for message in self.websocket:
                data = json.loads(message)
                self._handle_message(data)
        except websockets.ConnectionClosed:
            print("[System] Connection closed by server")
            self.connected = False
        except Exception as e:
            print(f"[Error] Receive error: {e}")
            self.connected = False

    def _handle_message(self, data: dict):
        msg_type = data.get("type", "unknown")
        timestamp = data.get("timestamp", "")
        nickname = data.get("nickname", "")

        if msg_type == "chat":
            content = data.get("content", "")
            print(f"[{timestamp}] {nickname}: {content}")

        elif msg_type == "join":
            print(f"[{timestamp}] [System] {nickname} joined the chat")

        elif msg_type == "leave":
            print(f"[{timestamp}] [System] {nickname} left the chat")

        elif msg_type == "nick_change":
            new_nickname = data.get("new_nickname", "")
            print(f"[{timestamp}] [System] {nickname} is now known as {new_nickname}")

        elif msg_type == "error":
            content = data.get("content", "")
            print(f"[{timestamp}] [Error] {content}")

        else:
            print(f"[{timestamp}] [Unknown] {data}")


async def interactive_client():
    print("=" * 50)
    print("WebSocket Chat Client")
    print("=" * 50)
    print("Commands:")
    print("  /nick <name>  - Change your nickname")
    print("  /quit         - Disconnect and exit")
    print("  /help         - Show this help")
    print("  <message>     - Send a chat message")
    print("=" * 50)

    server_url = input("Enter server URL (default: ws://localhost:8080/ws): ").strip()
    if not server_url:
        server_url = "ws://localhost:8080/ws"

    nickname = input("Enter your nickname (default: Anonymous): ").strip()
    if not nickname:
        nickname = "Anonymous"

    client = ChatClient(server_url)

    if not await client.connect(nickname):
        print("Failed to connect. Exiting.")
        return

    receive_task = asyncio.create_task(client.receive_messages())

    print("\n--- Chat started ---")

    try:
        while client.connected:
            try:
                user_input = await asyncio.get_event_loop().run_in_executor(
                    None, lambda: input()
                )
            except EOFError:
                break

            if not user_input:
                continue

            if user_input.startswith("/"):
                parts = user_input.split(maxsplit=1)
                command = parts[0].lower()

                if command == "/quit":
                    break
                elif command == "/nick":
                    if len(parts) > 1:
                        await client.change_nickname(parts[1])
                        client.nickname = parts[1]
                    else:
                        print("Usage: /nick <new_nickname>")
                elif command == "/help":
                    print("\nCommands:")
                    print("  /nick <name>  - Change your nickname")
                    print("  /quit         - Disconnect and exit")
                    print("  /help         - Show this help")
                    print("  <message>     - Send a chat message\n")
                else:
                    print(f"Unknown command: {command}. Type /help for commands.")
            else:
                await client.send_message(user_input)

    except KeyboardInterrupt:
        print("\n[Interrupted]")
    finally:
        receive_task.cancel()
        await client.disconnect()


async def demo_client():
    print("=" * 50)
    print("WebSocket Chat Client - Demo Mode")
    print("=" * 50)

    client = ChatClient()

    if not await client.connect("DemoUser"):
        return

    await client.send_message("Hello from Python client!")
    await asyncio.sleep(1)

    await client.send_message("This is a demo message")
    await asyncio.sleep(1)

    await client.change_nickname("NewDemoUser")
    await asyncio.sleep(1)

    await client.send_message("I changed my nickname!")
    await asyncio.sleep(2)

    await client.disconnect()


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--demo":
        asyncio.run(demo_client())
    else:
        asyncio.run(interactive_client())
