from flask import Flask, jsonify, request

app = Flask(__name__)

ORDERS = {
    1: [
        {"order_id": 101, "product": "Ноутбук", "price": 75000, "status": "delivered"},
        {"order_id": 102, "product": "Мышь", "price": 2500, "status": "shipped"},
    ],
    2: [
        {"order_id": 201, "product": "Клавиатура", "price": 5000, "status": "processing"},
    ],
    3: [
        {"order_id": 301, "product": "Монитор", "price": 25000, "status": "delivered"},
        {"order_id": 302, "product": "Веб-камера", "price": 3500, "status": "cancelled"},
        {"order_id": 303, "product": "Наушники", "price": 4000, "status": "shipped"},
    ],
}


@app.route("/orders", methods=["GET"])
def get_orders():
    user_id = request.args.get("user_id")
    
    if not user_id:
        return jsonify({"error": "user_id is required"}), 400
    
    try:
        user_id = int(user_id)
    except ValueError:
        return jsonify({"error": "invalid user_id"}), 400
    
    orders = ORDERS.get(user_id)
    
    if orders is None:
        return jsonify({"error": "user not found"}), 404
    
    return jsonify({
        "user_id": user_id,
        "orders": orders
    })


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8082)
