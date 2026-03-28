import pytest
import requests
import time

BASE_URL = "http://localhost:8000"


@pytest.fixture(scope="module", autouse=True)
def wait_for_server():
    max_retries = 10
    for i in range(max_retries):
        try:
            response = requests.get(f"{BASE_URL}/health", timeout=2)
            if response.status_code == 200:
                return
        except requests.exceptions.RequestException:
            pass
        time.sleep(1)
    pytest.fail("Server did not start within timeout period")


@pytest.fixture(autouse=True)
def cleanup_items():
    response = requests.get(f"{BASE_URL}/items")
    for item in response.json():
        requests.delete(f"{BASE_URL}/items/{item["id"]}")
    yield
    response = requests.get(f"{BASE_URL}/items")
    for item in response.json():
        requests.delete(f"{BASE_URL}/items/{item["id"]}")


class TestRoot:

    def test_root_returns_welcome_message(self):
        response = requests.get(f"{BASE_URL}/")
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "Welcome to Task3 API"


class TestHealthCheck:

    def test_health_returns_healthy_status(self):
        response = requests.get(f"{BASE_URL}/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"


class TestEcho:

    def test_echo_returns_message_with_length(self):
        payload = {"message": "Hello World"}
        response = requests.post(f"{BASE_URL}/echo", json=payload)
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "Hello World"
        assert data["length"] == 11

    def test_echo_empty_message(self):
        payload = {"message": ""}
        response = requests.post(f"{BASE_URL}/echo", json=payload)
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == ""
        assert data["length"] == 0

    def test_echo_long_message(self):
        payload = {"message": "A" * 1000}
        response = requests.post(f"{BASE_URL}/echo", json=payload)
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "A" * 1000
        assert data["length"] == 1000


class TestItems:

    def test_get_items_empty(self):
        response = requests.get(f"{BASE_URL}/items")
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)
        assert len(data) == 0

    def test_create_item(self):
        payload = {
            "name": "Test Item",
            "description": "Test Description",
            "price": 19.99
        }
        response = requests.post(f"{BASE_URL}/items", json=payload)
        assert response.status_code == 201
        data = response.json()
        assert data["id"] > 0
        assert data["name"] == "Test Item"
        assert data["description"] == "Test Description"
        assert data["price"] == 19.99

    def test_create_item_without_description(self):
        payload = {
            "name": "Test Item",
            "price": 29.99
        }
        response = requests.post(f"{BASE_URL}/items", json=payload)
        assert response.status_code == 201
        data = response.json()
        assert data["id"] > 0
        assert data["name"] == "Test Item"
        assert data["description"] is None
        assert data["price"] == 29.99

    def test_get_item_by_id(self):
        payload = {"name": "Item 1", "price": 10.0}
        create_response = requests.post(f"{BASE_URL}/items", json=payload)
        item_id = create_response.json()["id"]

        response = requests.get(f"{BASE_URL}/items/{item_id}")
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == item_id
        assert data["name"] == "Item 1"
        assert data["price"] == 10.0

    def test_get_item_not_found(self):
        response = requests.get(f"{BASE_URL}/items/9999")
        assert response.status_code == 404

    def test_update_item(self):
        payload = {"name": "Original", "price": 10.0}
        create_response = requests.post(f"{BASE_URL}/items", json=payload)
        item_id = create_response.json()["id"]

        update_payload = {"name": "Updated", "price": 20.0}
        response = requests.put(f"{BASE_URL}/items/{item_id}", json=update_payload)
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == item_id
        assert data["name"] == "Updated"
        assert data["price"] == 20.0

    def test_update_item_not_found(self):
        payload = {"name": "Test", "price": 10.0}
        response = requests.put(f"{BASE_URL}/items/9999", json=payload)
        assert response.status_code == 404

    def test_delete_item(self):
        payload = {"name": "To Delete", "price": 5.0}
        create_response = requests.post(f"{BASE_URL}/items", json=payload)
        item_id = create_response.json()["id"]

        response = requests.delete(f"{BASE_URL}/items/{item_id}")
        assert response.status_code == 204

        get_response = requests.get(f"{BASE_URL}/items/{item_id}")
        assert get_response.status_code == 404

    def test_delete_item_not_found(self):
        response = requests.delete(f"{BASE_URL}/items/9999")
        assert response.status_code == 404

    def test_get_multiple_items(self):
        for i in range(3):
            payload = {"name": f"Item {i}", "price": float(i) + 0.01}
            requests.post(f"{BASE_URL}/items", json=payload)

        response = requests.get(f"{BASE_URL}/items")
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 3


class TestItemValidation:

    def test_create_item_with_zero_price(self):
        payload = {"name": "Free Item", "price": 0.0}
        response = requests.post(f"{BASE_URL}/items", json=payload)
        assert response.status_code == 422

    def test_create_item_with_negative_price(self):
        payload = {"name": "Negative Item", "price": -10.0}
        response = requests.post(f"{BASE_URL}/items", json=payload)
        assert response.status_code == 422

    def test_create_item_missing_name(self):
        payload = {"price": 10.0}
        response = requests.post(f"{BASE_URL}/items", json=payload)
        assert response.status_code == 422

    def test_create_item_empty_name(self):
        payload = {"name": "", "price": 10.0}
        response = requests.post(f"{BASE_URL}/items", json=payload)
        assert response.status_code == 422

    def test_create_item_whitespace_name(self):
        payload = {"name": "   ", "price": 10.0}
        response = requests.post(f"{BASE_URL}/items", json=payload)
        assert response.status_code == 422
