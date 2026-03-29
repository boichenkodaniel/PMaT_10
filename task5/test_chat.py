#!/usr/bin/env python3

import asyncio
import json
import pytest
import websockets
from typing import Optional

SERVER_URL = "ws://localhost:8080/ws"
CLIENTS_URL = "http://localhost:8080/clients"


class ChatClient:

    def __init__(self, nickname: str = "TestUser"):
        self.nickname = nickname
        self.websocket: Optional[websockets.WebSocketClientProtocol] = None
        self.received_messages = []
        self.connected = False

    async def connect(self, url: str = SERVER_URL) -> bool:
        try:
            self.websocket = await websockets.connect(url)
            self.connected = True
            await self.change_nickname(self.nickname)
            return True
        except Exception:
            return False

    async def disconnect(self):
        if self.websocket:
            await self.websocket.close()
            self.connected = False

    async def send_message(self, content: str):
        if self.websocket:
            msg = {"type": "chat", "content": content}
            await self.websocket.send(json.dumps(msg))

    async def change_nickname(self, new_nickname: str):
        if self.websocket:
            msg = {"type": "nick_change", "new_nickname": new_nickname}
            await self.websocket.send(json.dumps(msg))

    async def receive_message(self, timeout: float = 2.0) -> Optional[dict]:
        if not self.websocket:
            return None

        try:
            message = await asyncio.wait_for(
                self.websocket.recv(),
                timeout=timeout
            )
            data = json.loads(message)
            self.received_messages.append(data)
            return data
        except asyncio.TimeoutError:
            return None
        except websockets.ConnectionClosed:
            self.connected = False
            return None

    async def receive_until_type(self, msg_type: str, timeout: float = 2.0) -> Optional[dict]:
        start_time = asyncio.get_event_loop().time()
        while asyncio.get_event_loop().time() - start_time < timeout:
            msg = await self.receive_message(timeout=0.5)
            if msg and msg.get("type") == msg_type:
                return msg
        return None


async def check_server():
    try:
        ws = await websockets.connect(SERVER_URL, close_timeout=1)
        await ws.close()
        return True
    except Exception:
        return False


@pytest.fixture(scope="module")
def event_loop():
    loop = asyncio.new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
async def client():
    c = ChatClient("TestUser")
    connected = await c.connect()
    if not connected:
        pytest.skip("Server not available")
    yield c
    await c.disconnect()


@pytest.fixture
async def two_clients():
    c1 = ChatClient("User1")
    c2 = ChatClient("User2")

    connected1 = await c1.connect()
    connected2 = await c2.connect()

    if not (connected1 and connected2):
        pytest.skip("Server not available")

    yield c1, c2

    await c1.disconnect()
    await c2.disconnect()


class TestConnection:

    @pytest.mark.asyncio
    async def test_connect_to_server(self):
        client = ChatClient("ConnectTest")
        result = await client.connect()
        assert result is True
        assert client.connected is True
        await client.disconnect()

    @pytest.mark.asyncio
    async def test_multiple_clients_connect(self):
        clients = []
        for i in range(3):
            c = ChatClient(f"Client{i}")
            result = await c.connect()
            assert result is True
            clients.append(c)

        for c in clients:
            await c.disconnect()

    @pytest.mark.asyncio
    async def test_disconnect(self):
        client = ChatClient("DisconnectTest")
        await client.connect()
        assert client.connected is True

        await client.disconnect()
        assert client.connected is False


class TestMessaging:

    @pytest.mark.asyncio
    async def test_send_message(self, client):
        await client.send_message("Hello, World!")
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert msg["type"] == "chat"
        assert msg["content"] == "Hello, World!"

    @pytest.mark.asyncio
    async def test_message_contains_nickname(self, client):
        await client.send_message("Test message")
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert "nickname" in msg
        assert msg["nickname"] == "TestUser"

    @pytest.mark.asyncio
    async def test_message_contains_timestamp(self, client):
        await client.send_message("Timestamp test")
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert "timestamp" in msg
        assert msg["timestamp"] is not None

    @pytest.mark.asyncio
    async def test_broadcast_to_all_clients(self, two_clients):
        client1, client2 = two_clients

        await client1.send_message("Broadcast test")

        msg1 = await client1.receive_until_type("chat", timeout=2.0)
        msg2 = await client2.receive_until_type("chat", timeout=2.0)

        assert msg1 is not None
        assert msg2 is not None
        assert msg1["content"] == "Broadcast test"
        assert msg2["content"] == "Broadcast test"


class TestNicknames:

    @pytest.mark.asyncio
    async def test_change_nickname(self, client):
        while True:
            msg = await client.receive_message(timeout=0.3)
            if msg is None:
                break

        await client.change_nickname("NewNickname")
        msg = await client.receive_until_type("nick_change", timeout=2.0)

        assert msg is not None
        assert msg["type"] == "nick_change"
        assert msg["new_nickname"] == "NewNickname"

    @pytest.mark.asyncio
    async def test_nickname_change_broadcast(self, two_clients):
        client1, client2 = two_clients

        while True:
            msg = await client1.receive_message(timeout=0.3)
            if msg is None:
                break

        while True:
            msg = await client2.receive_message(timeout=0.3)
            if msg is None:
                break

        await client1.change_nickname("ChangedName")

        msg = await client2.receive_until_type("nick_change", timeout=2.0)

        assert msg is not None
        assert msg["type"] == "nick_change"
        assert msg["new_nickname"] == "ChangedName"

    @pytest.mark.asyncio
    async def test_messages_use_new_nickname(self, client):
        while True:
            msg = await client.receive_message(timeout=0.3)
            if msg is None:
                break

        await client.change_nickname("UpdatedUser")
        await client.receive_until_type("nick_change", timeout=2.0)

        await client.send_message("Using new name")
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert msg["nickname"] == "UpdatedUser"


class TestNotifications:

    @pytest.mark.asyncio
    async def test_join_notification(self, two_clients):
        client1, client2 = two_clients

        msg = await client2.receive_until_type("join", timeout=2.0)

        assert msg is not None
        assert msg["type"] == "join"
        assert "nickname" in msg

    @pytest.mark.asyncio
    async def test_leave_notification(self):
        client1 = ChatClient("LeaveTest1")
        client2 = ChatClient("LeaveTest2")

        await client1.connect()
        await client2.connect()

        await asyncio.sleep(0.5)

        await client1.disconnect()

        msg = await client2.receive_until_type("leave", timeout=2.0)

        assert msg is not None
        assert msg["type"] == "leave"
        assert msg["nickname"] == "LeaveTest1"

        await client2.disconnect()


class TestMessageTypes:

    @pytest.mark.asyncio
    async def test_chat_message_type(self, client):
        await client.send_message("Test")
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg["type"] == "chat"

    @pytest.mark.asyncio
    async def test_join_message_type(self, client):
        new_client = ChatClient("JoinTypeTest")
        await new_client.connect()

        msg = await client.receive_until_type("join", timeout=2.0)

        assert msg is not None
        assert msg["type"] == "join"
        assert "nickname" in msg
        assert "timestamp" in msg

        await new_client.disconnect()

    @pytest.mark.asyncio
    async def test_leave_message_type(self):
        client1 = ChatClient("LeaveType1")
        client2 = ChatClient("LeaveType2")

        await client1.connect()
        await client2.connect()
        await asyncio.sleep(0.3)

        await client1.disconnect()

        msg = await client2.receive_until_type("leave", timeout=2.0)

        assert msg is not None
        assert msg["type"] == "leave"
        assert "nickname" in msg
        assert "timestamp" in msg

        await client2.disconnect()


class TestMessageFormat:

    @pytest.mark.asyncio
    async def test_message_is_valid_json(self, client):
        await client.send_message("JSON test")
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert isinstance(msg, dict)

    @pytest.mark.asyncio
    async def test_timestamp_format(self, client):
        await client.send_message("Time test")
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert "timestamp" in msg
        timestamp = msg["timestamp"]
        assert "T" in timestamp
        assert len(timestamp) > 10


class TestErrorHandling:

    @pytest.mark.asyncio
    async def test_empty_message(self, client):
        while True:
            msg = await client.receive_message(timeout=0.3)
            if msg is None:
                break

        await client.send_message("")
        msg = await client.receive_message(timeout=1.0)
        if msg:
            assert msg["type"] in ["chat", "error", "join", "leave", "nick_change"]

    @pytest.mark.asyncio
    async def test_long_message(self, client):
        long_content = "A" * 1000
        await client.send_message(long_content)
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert msg["content"] == long_content

    @pytest.mark.asyncio
    async def test_special_characters_in_message(self, client):
        special_msg = "Test with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"
        await client.send_message(special_msg)
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert msg["content"] == special_msg

    @pytest.mark.asyncio
    async def test_unicode_in_message(self, client):
        unicode_msg = "Hello 世界 🌍 Привет"
        await client.send_message(unicode_msg)
        msg = await client.receive_until_type("chat", timeout=2.0)

        assert msg is not None
        assert msg["content"] == unicode_msg


class TestConcurrency:

    @pytest.mark.asyncio
    async def test_concurrent_messages(self, two_clients):
        client1, client2 = two_clients

        for i in range(5):
            await client1.send_message(f"Message {i}")

        received = 0
        for _ in range(5):
            msg = await client2.receive_until_type("chat", timeout=2.0)
            if msg:
                received += 1

        assert received == 5

    @pytest.mark.asyncio
    async def test_rapid_connect_disconnect(self):
        for i in range(3):
            client = ChatClient(f"RapidUser{i}")
            await client.connect()
            await asyncio.sleep(0.1)
            await client.disconnect()
            await asyncio.sleep(0.1)


@pytest.mark.asyncio
async def test_server_available():
    available = await check_server()
    if not available:
        pytest.skip("Chat server is not running. Start it with: go run main.go")
    assert available is True


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
