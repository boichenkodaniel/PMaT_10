"""
Pytest tests for Order Service and API Gateway Integration
"""
import pytest
import requests
import subprocess
import time
import signal
import os
import sys


GATEWAY_URL = "http://localhost:8080"
USER_SERVICE_URL = "http://localhost:8081"
ORDER_SERVICE_URL = "http://localhost:8082"


@pytest.fixture
def order_client():
    order_svc_path = os.path.join(os.path.dirname(__file__), "order-service")
    sys.path.insert(0, order_svc_path)
    from app import app
    app.config['TESTING'] = True
    with app.test_client() as client:
        yield client


class TestOrderService:

    def test_get_orders_success(self, order_client):
        response = order_client.get('/orders?user_id=1')
        data = response.get_json()

        assert response.status_code == 200
        assert data['user_id'] == 1
        assert isinstance(data['orders'], list)
        assert len(data['orders']) > 0

    def test_get_orders_user_2(self, order_client):
        response = order_client.get('/orders?user_id=2')
        data = response.get_json()

        assert response.status_code == 200
        assert data['user_id'] == 2
        assert len(data['orders']) == 1

    def test_get_orders_user_3(self, order_client):
        response = order_client.get('/orders?user_id=3')
        data = response.get_json()

        assert response.status_code == 200
        assert data['user_id'] == 3
        assert len(data['orders']) == 3

    def test_get_orders_missing_user_id(self, order_client):
        response = order_client.get('/orders')
        data = response.get_json()

        assert response.status_code == 400
        assert 'error' in data
        assert 'required' in data['error']

    def test_get_orders_invalid_user_id(self, order_client):
        response = order_client.get('/orders?user_id=abc')
        data = response.get_json()

        assert response.status_code == 400
        assert 'error' in data
        assert 'invalid' in data['error']

    def test_get_orders_user_not_found(self, order_client):
        response = order_client.get('/orders?user_id=999')
        data = response.get_json()

        assert response.status_code == 404
        assert 'error' in data
        assert 'not found' in data['error']

    def test_order_structure(self, order_client):
        response = order_client.get('/orders?user_id=1')
        data = response.get_json()

        order = data['orders'][0]
        assert 'order_id' in order
        assert 'product' in order
        assert 'price' in order
        assert 'status' in order

    def test_order_field_types(self, order_client):
        response = order_client.get('/orders?user_id=1')
        data = response.get_json()

        order = data['orders'][0]
        assert isinstance(order['order_id'], int)
        assert isinstance(order['product'], str)
        assert isinstance(order['price'], (int, float))
        assert isinstance(order['status'], str)


class ServiceManager:
    def __init__(self):
        self.processes = []

    def start_service(self, cmd, cwd=None):
        try:
            proc = subprocess.Popen(
                cmd,
                cwd=cwd,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                creationflags=subprocess.CREATE_NEW_PROCESS_GROUP if sys.platform == 'win32' else 0
            )
            self.processes.append(proc)
            return proc
        except Exception as e:
            print(f"Failed to start service: {e}")
            return None

    def stop_all(self):
        for proc in self.processes:
            try:
                if sys.platform == 'win32':
                    proc.terminate()
                else:
                    proc.send_signal(signal.SIGTERM)
            except:
                pass
        self.processes.clear()


@pytest.fixture(scope="module")
def services():
    manager = ServiceManager()

    user_svc_path = os.path.join(os.path.dirname(__file__), "user-service")
    manager.start_service(["go", "run", "main.go"], cwd=user_svc_path)

    order_svc_path = os.path.join(os.path.dirname(__file__), "order-service")
    manager.start_service([sys.executable, "app.py"], cwd=order_svc_path)

    gateway_path = os.path.join(os.path.dirname(__file__), "gateway")
    manager.start_service(["go", "run", "main.go"], cwd=gateway_path)

    time.sleep(3)

    yield manager

    manager.stop_all()


@pytest.mark.usefixtures("services")
class TestGatewayIntegration:
    def test_gateway_health(self):
        response = requests.get(f"{GATEWAY_URL}/health", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"

    def test_gateway_root(self):
        response = requests.get(f"{GATEWAY_URL}/", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["name"] == "API Gateway"

    def test_user_service_direct(self):
        response = requests.get(f"{USER_SERVICE_URL}/user?id=1", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == 1
        assert "name" in data
        assert "phone" in data

    def test_order_service_direct(self):
        response = requests.get(f"{ORDER_SERVICE_URL}/orders?user_id=1", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["user_id"] == 1
        assert "orders" in data

    def test_gateway_user_endpoint(self):
        response = requests.get(f"{GATEWAY_URL}/api/user?id=1", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == 1
        assert "name" in data
        assert "phone" in data

    def test_gateway_orders_endpoint(self):
        response = requests.get(f"{GATEWAY_URL}/api/orders?user_id=1", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["user_id"] == 1
        assert "orders" in data
        assert isinstance(data["orders"], list)

    def test_gateway_profile_endpoint(self):
        response = requests.get(f"{GATEWAY_URL}/api/profile?id=1", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert "user" in data
        assert "orders" in data
        assert data["user"]["id"] == 1
        assert isinstance(data["orders"], list)

    def test_gateway_profile_user_2(self):
        response = requests.get(f"{GATEWAY_URL}/api/profile?id=2", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["user"]["id"] == 2
        assert len(data["orders"]) >= 1

    def test_gateway_user_not_found(self):
        response = requests.get(f"{GATEWAY_URL}/api/user?id=999", timeout=5)
        assert response.status_code in [404, 503]

    def test_gateway_missing_user_id(self):
        response = requests.get(f"{GATEWAY_URL}/api/user", timeout=5)
        assert response.status_code == 400

    def test_gateway_orders_missing_user_id(self):
        response = requests.get(f"{GATEWAY_URL}/api/orders", timeout=5)
        assert response.status_code in [400, 503]

    def test_gateway_profile_missing_id(self):
        response = requests.get(f"{GATEWAY_URL}/api/profile", timeout=5)
        assert response.status_code == 400


@pytest.mark.usefixtures("services")
class TestAllUsers:
    @pytest.mark.parametrize("user_id", [1, 2, 3])
    def test_all_users_profile(self, user_id):
        response = requests.get(f"{GATEWAY_URL}/api/profile?id={user_id}", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["user"]["id"] == user_id
        assert "orders" in data


@pytest.mark.usefixtures("services")
class TestServiceHealth:
    def test_user_service_health_via_gateway(self):
        response = requests.get(f"{GATEWAY_URL}/health", timeout=5)
        data = response.json()
        assert data["user_service"] == "up"

    def test_order_service_health_via_gateway(self):
        response = requests.get(f"{GATEWAY_URL}/health", timeout=5)
        data = response.json()
        assert data["order_service"] == "up"


if __name__ == '__main__':
    pytest.main([__file__, '-v', '-s'])
