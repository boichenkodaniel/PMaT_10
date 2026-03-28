import pytest
import requests
import time
import subprocess
import os
import signal

SERVER_URL = "http://localhost:8080"
server_process = None


@pytest.fixture(scope="module", autouse=True)
def start_server():
    global server_process

    server_process = subprocess.Popen(
        ["go", "run", "main.go"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=os.path.dirname(os.path.abspath(__file__))
    )

    time.sleep(2)

    yield

    if server_process:
        server_process.send_signal(signal.SIGTERM)
        server_process.wait(timeout=5)


def test_root_endpoint():
    response = requests.get(SERVER_URL, timeout=5)
    assert response.status_code == 200
    assert response.text == "Hello, World!"


def test_health_endpoint():
    response = requests.get(f"{SERVER_URL}/health", timeout=5)
    assert response.status_code == 200
    assert response.text == "OK"


def test_logging_middleware():
    response = requests.get(f"{SERVER_URL}/test-path", timeout=5)
    assert response.status_code == 200


def test_unknown_path():
    response = requests.get(f"{SERVER_URL}/unknown", timeout=5)
    assert response.status_code == 200
