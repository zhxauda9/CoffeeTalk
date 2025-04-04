async function loadOrders() {
    try {
        let url = `/orders`;
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Orders could not be loaded.');
        }
        const orders = await response.json();
        const table = document.getElementById('order-table');
        table.innerHTML = '';
        orders.forEach(order => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${order.order_id}</td>
                <td><input type="text" class="form-control" value='${order.customer_name}' id="customer-name-${order.order_id}"></td>
                <td><textarea class="form-control" id="items-${order.order_id}">${JSON.stringify(order.items)}</textarea></td>
                <td>
                    <select class="form-control" id="status-${order.order_id}">
                        <option value="open" ${order.status === "open" ? "selected" : ""}>Open</option>
                        <option value="closed" ${order.status === "closed" ? "selected" : ""}>Closed</option>
                    </select>
                </td>
                <td><textarea class="form-control" id="notes-${order.order_id}">${JSON.stringify(order.notes)}</textarea></td>
                <td>${order.created_at}</td>
                <td>
                    <button class="btn btn-success btn-sm" onclick='closeOrder(${order.order_id})'>Close</button>
                    <button class="btn btn-warning btn-sm" onclick='updateOrder(${order.order_id})'>Update</button>
                    <button class="btn btn-danger btn-sm" onclick='deleteOrder(${order.order_id})'>Delete</button>
                </td>
            `;
            table.appendChild(row);
        });
    } catch (error) {
        console.error(error);
        alert("Orders loading error.");
    }
}

loadOrders();

document.getElementById('create-order-form').addEventListener('submit', async (e) => {
    e.preventDefault();

    const customerName = document.getElementById("customer-name").value;
    const itemsInput = document.getElementById("order-items").value.trim();
    const notesInput = document.getElementById("order-notes").value.trim();
    const status = document.getElementById("order-status").value;

    let items, notes;
    try {
        items = JSON.parse(itemsInput);
        if (!Array.isArray(items)) throw new Error("Items should be an array.");
    } catch (error) {
        alert("Invalid format for items. Must be JSON array.");
        return;
    }

    try {
        notes = notesInput ? JSON.parse(notesInput) : {};
    } catch (error) {
        alert("Invalid format for notes. Must be valid JSON.");
        return;
    }

    const order = {
        customer_name: customerName,
        items: items,
        status: status,
        notes: notes
    };

    try {
        const response = await fetch(`/orders`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(order),
        });

        if (!response.ok) {
            throw new Error("Failed to create order.");
        }

        alert("Order created successfully!");
        loadOrders();
    } catch (error) {
        console.error(error);
        alert("Error: Failed to create order.");
    }
});


async function updateOrder(id) {
    const name = document.getElementById(`customer-name-${id}`).value;
    const status = document.getElementById(`status-${id}`).value;

    // Обработка items
    const itemsText = document.getElementById(`items-${id}`)?.value;
    let items = [];
    try {
        items = JSON.parse(itemsText);
        if (!Array.isArray(items)) throw new Error("Items should be an array.");
    } catch (error) {
        alert("Invalid format for items. Must be JSON array.");
        return;
    }

    // Обработка notes
    const notesText = document.getElementById(`notes-${id}`)?.value;
    let notes = null; // Значение по умолчанию - null, если заметки пустые
    if (notesText) {
        try {
            notes = JSON.parse(notesText); // Парсим заметки, если они не пустые
        } catch (error) {
            alert("Invalid format for notes. Must be valid JSON.");
            return;
        }
    }

    const order = {
        customer_name: name,
        status: status,
        items: items,
        notes: notes
    };

    try {
        const response = await fetch(`/orders/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(order),
        });

        if (!response.ok) {
            throw new Error(`Failed to update order: ${id}.`);
        }

        alert("Order updated successfully!");
        loadOrders();
    } catch (error) {
        console.error(error);
        alert(`Error: Failed to update order: ${id}.`);
    }
}



async function deleteOrder(id) {
    if (confirm("Are you sure you want to delete the order?")) {
        try {
            const response = await fetch(`/orders/${id}`, { method: 'DELETE' });

            if (!response.ok) {
                throw new Error(`Failed to delete order: ${id}.`);
            }

            alert(`Order ${id} deleted.`);
            loadOrders();
        } catch (error) {
            console.error(error);
            alert(`Error deleting order: ${id}.`);
        }
    }
}

async function closeOrder(id) {
    try {
        const response = await fetch(`/orders/${id}/close`, { method: 'POST' });

        if (!response.ok) {
            throw new Error(`Failed to close order: ${id}.`);
        }

        alert(`Order ${id} closed.`);
        loadOrders();
    } catch (error) {
        console.error(error);
        alert(`Error closing order: ${id}.`);
    }
}
