async function loadInventory() {
    try {
        let url = `/inventory`;
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Inventory could not be loaded.');
        }
        const inventory = await response.json();
        const table = document.getElementById('inventory-table');
        table.innerHTML = '';

        inventory.forEach(item => {
            const row = document.createElement('tr');
            row.innerHTML = `
            <td>${item.ingredient_id}</td>
            <td><input type="text" class="form-control" value="${item.name}" id="name-${item.ingredient_id}"></td>
            <td><input type="number" class="form-control" value="${item.quantity}" id="quantity-${item.ingredient_id}"></td>
            <td><input type="text" class="form-control" value="${item.unit}" id="unit-${item.ingredient_id}"></td>
            <td>
                <button class="btn btn-warning btn-sm" onclick="updateInventory(${item.ingredient_id})">Update</button>
                <button class="btn btn-danger btn-sm" onclick="deleteInventory(${item.ingredient_id})">Delete</button>
            </td>
            `;
            table.appendChild(row);
        });
    } catch (error) {
        console.error(error);
        alert("Inventory loading error.");
    }
}

async function updateInventory(id) {
    const name = document.getElementById(`name-${id}`).value;
    const quantity = document.getElementById(`quantity-${id}`).value;
    const unit = document.getElementById(`unit-${id}`).value;

    const item = {
        ingredient_id: id,
        name: name,
        quantity: parseFloat(quantity),
        unit: unit,
    };

    try {
        const response = await fetch(`/inventory/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item),
        });

        if (!response.ok) {
            throw new Error(`Failed to update inventory: ${id}.`);
        }

        alert("Inventory item updated successfully!");
        loadInventory();
    } catch (error) {
        console.error(error);
        alert(`Error: Failed to update inventory item: ${id}.`);
    }
}

async function deleteInventory(id) {
    if (confirm("Are you sure you want to delete the inventory item?")) {
        try {
            const response = await fetch(`/inventory/${id}`, { method: 'DELETE' });

            if (!response.ok) {
                throw new Error(`Failed to delete inventory item: ${id}.`);
            }

            alert("Inventory deleted.");
            loadInventory();
        } catch (error) {
            console.error(error);
            alert(`Error deleting inventory item: ${id}.`);
        }
    }
}

document.getElementById('create-inventory-form').addEventListener('submit', async (e) => {
    e.preventDefault();

    const name = document.getElementById('ingredient-name').value;
    const quantity = document.getElementById('ingredient-quantity').value;
    const unit = document.getElementById('ingredient-unit').value;

    if (!name || !quantity || !unit) {
        alert("Please provide valid input data.");
        return;
    }

    const item = {
        name: name,
        quantity: parseFloat(quantity),
        unit: unit,
    };

    try {
        const response = await fetch(`/inventory`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item),
        });

        if (!response.ok) {
            throw new Error("Failed to create inventory item.");
        }

        alert("Inventory item created successfully!");
        loadInventory();
    } catch (error) {
        console.error(error);
        alert("Error: Failed to create inventory item.");
    }
});

loadInventory();