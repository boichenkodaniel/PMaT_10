from fastapi import FastAPI, HTTPException
import httpx
from pydantic import BaseModel

app = FastAPI()

GO_SERVICE_URL = "http://localhost:8080"


class MathRequest(BaseModel):
    a: int
    b: int


class MathResponse(BaseModel):
    result: int
    operation: str


@app.get("/health")
async def health():
    async with httpx.AsyncClient() as client:
        try:
            response = await client.get(f"{GO_SERVICE_URL}/health")
            go_health = response.json()
        except Exception:
            go_health = {"status": "unavailable"}

        return {
            "fastapi": "ok",
            "go_service": go_health
        }


@app.post("/sum", response_model=MathResponse)
async def sum_numbers(req: MathRequest):
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(
                f"{GO_SERVICE_URL}/sum",
                json={"a": req.a, "b": req.b}
            )
            if response.status_code != 200:
                raise HTTPException(status_code=response.status_code, detail="Go service error")
            data = response.json()
            return MathResponse(result=data["result"], operation="sum")
        except httpx.RequestError:
            raise HTTPException(status_code=503, detail="Go service unavailable")


@app.post("/multiply", response_model=MathResponse)
async def multiply_numbers(req: MathRequest):
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(
                f"{GO_SERVICE_URL}/multiply",
                json={"a": req.a, "b": req.b}
            )
            if response.status_code != 200:
                raise HTTPException(status_code=response.status_code, detail="Go service error")
            data = response.json()
            return MathResponse(result=data["result"], operation="multiply")
        except httpx.RequestError:
            raise HTTPException(status_code=503, detail="Go service unavailable")


@app.get("/calculate")
async def calculate(a: int, b: int):
    async with httpx.AsyncClient() as client:
        try:
            sum_resp = await client.post(f"{GO_SERVICE_URL}/sum", json={"a": a, "b": b})
            mul_resp = await client.post(f"{GO_SERVICE_URL}/multiply", json={"a": a, "b": b})

            return {
                "input": {"a": a, "b": b},
                "sum": sum_resp.json()["result"],
                "multiply": mul_resp.json()["result"],
                "source": "Go Calculator Service"
            }
        except httpx.RequestError:
            raise HTTPException(status_code=503, detail="Go service unavailable")
