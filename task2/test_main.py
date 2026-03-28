import pytest
import httpx
from fastapi.testclient import TestClient
from main import app, GO_SERVICE_URL
from unittest.mock import AsyncMock, patch

pytest_plugins = ('pytest_asyncio',)


@pytest.fixture
def client():
    return TestClient(app)


@pytest.fixture
def go_client():
    return httpx.Client(base_url=GO_SERVICE_URL)


def _create_mock_response(status_code, json_data):
    response = httpx.Response(status_code=status_code, json=json_data)
    response._request = httpx.Request("GET", "http://test")
    return response


class TestFastAPIHealth:
    def test_health_fastapi_only(self, client):
        with patch('httpx.AsyncClient.get', new_callable=AsyncMock) as mock_get:
            mock_get.side_effect = httpx.ConnectError("Connection refused")

            response = client.get("/health")
            assert response.status_code == 200
            data = response.json()
            assert data["fastapi"] == "ok"
            assert data["go_service"]["status"] == "unavailable"


class TestFastAPISum:
    def test_sum_success(self, client):
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = _create_mock_response(200, {"result": 8})

            response = client.post("/sum", json={"a": 5, "b": 3})
            assert response.status_code == 200
            data = response.json()
            assert data["result"] == 8
            assert data["operation"] == "sum"

    def test_sum_service_unavailable(self, client):
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.side_effect = httpx.ConnectError("Connection refused")

            response = client.post("/sum", json={"a": 5, "b": 3})
            assert response.status_code == 503
            assert response.json()["detail"] == "Go service unavailable"

    def test_sum_go_error(self, client):
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = _create_mock_response(500, {})

            response = client.post("/sum", json={"a": 5, "b": 3})
            assert response.status_code == 500


class TestFastAPIMultiply:
    def test_multiply_success(self, client):
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.return_value = _create_mock_response(200, {"result": 15})

            response = client.post("/multiply", json={"a": 5, "b": 3})
            assert response.status_code == 200
            data = response.json()
            assert data["result"] == 15
            assert data["operation"] == "multiply"

    def test_multiply_service_unavailable(self, client):
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.side_effect = httpx.ConnectError("Connection refused")

            response = client.post("/multiply", json={"a": 5, "b": 3})
            assert response.status_code == 503


class TestFastAPICalculate:
    def test_calculate_success(self, client):
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.side_effect = [
                _create_mock_response(200, {"result": 15}),
                _create_mock_response(200, {"result": 50})
            ]

            response = client.get("/calculate", params={"a": 5, "b": 10})
            assert response.status_code == 200
            data = response.json()
            assert data["input"] == {"a": 5, "b": 10}
            assert data["sum"] == 15
            assert data["multiply"] == 50
            assert data["source"] == "Go Calculator Service"

    def test_calculate_service_unavailable(self, client):
        with patch('httpx.AsyncClient.post', new_callable=AsyncMock) as mock_post:
            mock_post.side_effect = httpx.ConnectError("Connection refused")

            response = client.get("/calculate", params={"a": 5, "b": 10})
            assert response.status_code == 503


class TestGoServiceIntegration:
    @pytest.mark.integration
    def test_go_health(self, go_client):
        try:
            response = go_client.get("/health")
            if response.status_code == 503:
                pytest.skip("Go service not running (returned 503)")
            assert response.status_code == 200
            data = response.json()
            assert data["status"] == "ok"
            assert data["service"] == "go-calculator"
        except httpx.ConnectError:
            pytest.skip("Go service not running")

    @pytest.mark.integration
    def test_go_sum(self, go_client):
        try:
            response = go_client.post("/sum", json={"a": 10, "b": 20})
            if response.status_code == 503:
                pytest.skip("Go service not running (returned 503)")
            assert response.status_code == 200
            data = response.json()
            assert data["result"] == 30
        except httpx.ConnectError:
            pytest.skip("Go service not running")

    @pytest.mark.integration
    def test_go_multiply(self, go_client):
        try:
            response = go_client.post("/multiply", json={"a": 6, "b": 7})
            if response.status_code == 503:
                pytest.skip("Go service not running (returned 503)")
            assert response.status_code == 200
            data = response.json()
            assert data["result"] == 42
        except httpx.ConnectError:
            pytest.skip("Go service not running")

    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_full_integration(self):
        try:
            async with httpx.AsyncClient(base_url=GO_SERVICE_URL) as go_client:
                health_resp = await go_client.get("/health")
                if health_resp.status_code != 200:
                    pytest.skip(f"Go service not running (returned {health_resp.status_code})")

                sum_resp = await go_client.post("/sum", json={"a": 3, "b": 4})
                mul_resp = await go_client.post("/multiply", json={"a": 3, "b": 4})

                assert sum_resp.status_code == 200
                assert mul_resp.status_code == 200
                assert sum_resp.json()["result"] == 7
                assert mul_resp.json()["result"] == 12
        except httpx.ConnectError:
            pytest.skip("Go service not running")
