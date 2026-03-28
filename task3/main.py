import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field, field_validator
from typing import Optional

app = FastAPI(title="Task3 API", version="1.0.0")

items_db: dict = {}
next_id: int = 1


class Item(BaseModel):
    id: Optional[int] = None
    name: str
    description: Optional[str] = None
    price: float = Field(..., gt=0)

    @field_validator("name")
    @classmethod
    def name_not_empty(cls, v: str) -> str:
        if not v.strip():
            raise ValueError("name cannot be empty or whitespace")
        return v


class ItemResponse(BaseModel):
    id: int
    name: str
    description: Optional[str] = None
    price: float


class EchoRequest(BaseModel):
    message: str


class EchoResponse(BaseModel):
    message: str
    length: int


@app.get("/")
async def root():
    return {"message": "Welcome to Task3 API"}


@app.get("/health")
async def health_check():
    return {"status": "healthy"}


@app.post("/echo", response_model=EchoResponse)
async def echo(request: EchoRequest):
    return EchoResponse(message=request.message, length=len(request.message))


@app.get("/items", response_model=list[ItemResponse])
async def get_items():
    return [
        ItemResponse(id=item.id, name=item.name, description=item.description, price=item.price)
        for item in items_db.values()
    ]


@app.post("/items", response_model=ItemResponse, status_code=201)
async def create_item(item: Item):
    global next_id
    new_id = next_id
    next_id += 1
    new_item = Item(id=new_id, name=item.name, description=item.description, price=item.price)
    items_db[new_id] = new_item
    return ItemResponse(id=new_id, name=item.name, description=item.description, price=item.price)


@app.get("/items/{item_id}", response_model=ItemResponse)
async def get_item(item_id: int):
    if item_id not in items_db:
        raise HTTPException(status_code=404, detail="Item not found")
    item = items_db[item_id]
    return ItemResponse(id=item.id, name=item.name, description=item.description, price=item.price)


@app.put("/items/{item_id}", response_model=ItemResponse)
async def update_item(item_id: int, item: Item):
    if item_id not in items_db:
        raise HTTPException(status_code=404, detail="Item not found")
    updated_item = Item(id=item_id, name=item.name, description=item.description, price=item.price)
    items_db[item_id] = updated_item
    return ItemResponse(id=updated_item.id, name=updated_item.name, description=updated_item.description, price=updated_item.price)


@app.delete("/items/{item_id}", status_code=204)
async def delete_item(item_id: int):
    if item_id not in items_db:
        raise HTTPException(status_code=404, detail="Item not found")
    del items_db[item_id]


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
