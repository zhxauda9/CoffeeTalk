# ☕ **Frappuchino - The Ultimate Coffee Shop Management System** 💖

Ever wondered how your favorite coffee shop juggles a rush of orders, makes sure they always have your favorite blend ready, and remembers that you love your coffee with an extra shot of espresso? Well, behind the scenes, they use smart management systems to make everything run smoothly. ☕💻

This project, **Frappuchino**, is a simplified version of those systems, and it gives you a chance to get hands-on with the magic behind the coffee counter. It’s designed to handle all the essential things a coffee shop needs to keep running smoothly. So, let’s make the coffee shop world even better with some cool tech! 💡

---

## ✨ Key Features of Frappuchino:

- **Manage Orders:** Create, update, close, and even delete customer orders with ease. You’ve got this! 👨‍🍳✨
- **Oversee Inventory:** Keep track of stock levels to prevent shortages and ensure every cup is fresh and perfect. 📦
- **Update the Menu:** Add new drinks, update prices, or add delicious new pastries to your offerings. 🍰☕
  
So, whether you’re brewing up a batch of lattes or tracking down that last packet of sugar, Hot Coffee helps you handle it all in style! 🎉

---

## 🌟 API Endpoints

### **Orders**

| Method | Endpoint            | Description                         | Response                     |
|--------|---------------------|-------------------------------------|------------------------------|
| POST   | `/orders`           | Creates a new order.               | 🎉 201 Created               |
| GET    | `/orders`           | Retrieves all orders.              | 😎 200 OK                    |
| GET    | `/orders/{id}`      | Retrieves a specific order by ID.  | 😄 200 OK                    |
| PUT    | `/orders/{id}`      | Updates an existing order.         | ✨ 200 OK                    |
| DELETE | `/orders/{id}`      | Deletes an order.                  | 💥 204 No Content           |
| POST   | `/orders/{id}/close` | Closes an open order.             | 💫 200 OK                    |

---

### **Menu Items**

| Method | Endpoint            | Description                         | Response                     |
|--------|---------------------|-------------------------------------|------------------------------|
| POST   | `/menu`             | Adds a new menu item.              | 🍰 201 Created               |
| GET    | `/menu`             | Retrieves all menu items.          | 📜 200 OK                    |
| GET    | `/menu/{id}`        | Retrieves a specific menu item.    | 🍽️ 200 OK                    |
| PUT    | `/menu/{id}`        | Updates an existing menu item.     | ✨ 200 OK                    |
| DELETE | `/menu/{id}`        | Deletes a menu item.               | 💥 204 No Content           |

---

### **Inventory**

| Method | Endpoint            | Description                         | Response                     |
|--------|---------------------|-------------------------------------|------------------------------|
| POST   | `/inventory`        | Adds a new inventory item.         | 🎉 201 Created               |
| GET    | `/inventory`        | Retrieves all inventory items.     | 💡 200 OK                    |
| GET    | `/inventory/{id}`   | Retrieves a specific inventory item. | 📦 200 OK                   |
| PUT    | `/inventory/{id}`   | Updates an inventory item.         | ✨ 200 OK                    |
| DELETE | `/inventory/{id}`   | Deletes an inventory item.         | 💥 204 No Content           |

---

### **Reports and Aggregations**  

| Method | Endpoint                  | Description                       | Response                     |
|--------|---------------------------|-----------------------------------|------------------------------|
| GET    | `/reports/total-sales`    | Retrieves total sales amount.     | 💰 200 OK                    |
| GET    | `/reports/popular-items`  | Retrieves a list of popular menu items. | 📊 200 OK                |

---

## 💌 Request Examples

### **Create/Update Order Request:**
```http
POST /orders
Content-Type: application/json

{
    "customer_name": "Tyler Derden",
    "items": [
        {
            "product_id": "latte",
            "quantity": 2
        },
        {
            "product_id": "muffin",
            "quantity": 1
        }
    ]
}
```

### **Add/Update Menu Item Request:**
```http
POST /menu
Content-Type: application/json

{
    "product_id": "latte",
    "name": "Caffe Latte",
    "description": "Espresso with steamed milk",
    "price": 3.5,
    "ingredients": [
        {
            "ingredient_id": "espresso_shot",
            "quantity": 1
        },
        {
            "ingredient_id": "milk",
            "quantity": 200
        }
    ]
}
```

### **Add/Update Inventory Item Request:**
```http
POST /inventory
Content-Type: application/json

{
    "ingredient_id": "espresso_shot",
    "name": "Espresso Shot",
    "quantity": 490,
    "unit": "shots"
}
```

---

### **Total Sales Aggregation Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "total_sales": 1500.50
}
```

---

## 🚨 Error Examples:

### **Invalid Product ID in Order Items:**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "Invalid product ID in order items."
}
```

### **Insufficient Inventory:**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "Insufficient inventory for ingredient 'Milk'. Required: 200ml, Available: 150ml."
}
```

---

## 📂 **Database Files** (Stored in the given directory — `--dir` flag)

1. **`orders.json`**  
   Contains information about all the orders placed in the coffee shop.

2. **`menu_items.json`**  
   Lists all the menu items, their descriptions, prices, and ingredients.

3. **`inventory.json`**  
   Tracks ingredients and their quantities, making sure the coffee shop is always stocked!

---

## 🏅 Authors:

This project has been brought to you with love by:
- **mboranba** 💻
- **azhalgas** 🌟
---

