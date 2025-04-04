async function loadMenu() {
    try {
        let url = `/menu`;
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Menu items could not be loaded.');
        }
        const items = await response.json();
        const table = document.getElementById('menu-table');
        table.innerHTML = '';
        items.forEach(item => {
            const row = document.createElement('tr');
            const imagePath = item.image.startsWith("uploads/") ? `/${item.image}` : `/uploads/${item.image}`;
            row.innerHTML = `
                <td>${item.product_id}</td>
                <td><input type="text" class="form-control" value='${item.name}' id="name-${item.product_id}"></td>
                <td><input type="text" class="form-control" value='${item.description}' id="description-${item.product_id}"></td>
                <td><input type="number" class="form-control" value="${item.price}" id="price-${item.product_id}"></td>
                <td><textarea class="form-control" id="ingredients-${item.product_id}">${JSON.stringify(item.ingredients)}</textarea></td>
                <td><img src="${imagePath}" alt='${item.product_id} image' style="width: 100px; height: auto;"></td>
                <td>
                    <input type="file" id='file-${item.product_id}' style="display:none;" onchange="changePhoto(${item.product_id})">
                    <button class="btn btn-success btn-sm" onclick='document.getElementById("file-${item.product_id}").click()'>Change</button>
                    <button class="btn btn-info btn-sm" onclick='removePhoto(${item.product_id})'>Remove</button>
                    <button class="btn btn-warning btn-sm" onclick='updateMenu(${item.product_id})'>Update</button>
                    <button class="btn btn-danger btn-sm" onclick='deleteMenu(${item.product_id})'>Delete</button>
                </td>
            `;
            table.appendChild(row);
        });
    } catch (error) {
        console.error(error);
        alert("Menu items loading error.");
    }
}

loadMenu();

document.getElementById('create-menu-form').addEventListener('submit', async (e) => {
    e.preventDefault();

    const menuName = document.getElementById("menu-name").value;
    const description = document.getElementById("menu-desc").value;
    const price = document.getElementById('menu-price').value;
    const ingsInput = document.getElementById("menu-ings").value.trim();
    const imageInput = document.getElementById('menu-image');

    let ingredients;
    try {
        ingredients = JSON.parse(ingsInput);
        if (!Array.isArray(ingredients)) throw new Error("Ingredients should be an array.");
    } catch (error) {
        alert("Invalid format for ingredients. Must be JSON array.");
        return;
    }

    const formData = new FormData();
    formData.append("name", menuName);
    formData.append("description", description);
    formData.append("price", price);
    formData.append("ingredients", JSON.stringify(ingredients));

    if (imageInput.files.length > 0) {
        formData.append("image", imageInput.files[0]);
    }

    try {
        const response = await fetch(`/menu`, {
            method: 'POST',
            body: formData,
        });

        if (!response.ok) {
            throw new Error("Failed to create menu item.");
        }

        alert("Menu item created successfully!");
        loadMenu();
    } catch (error) {
        console.error(error);
        alert("Error: Failed to create menu item.");
    }
});


async function changePhoto(id) {
    const fileInput = document.getElementById(`file-${id}`);
    if (!fileInput.files.length) {
        alert("No file selected.");
        return;
    }

    const formData = new FormData();
    formData.append("image", fileInput.files[0]);

    try {
        const response = await fetch(`/menu/${id}/image`, {
            method: 'PUT',
            body: formData
        });

        if (!response.ok) {
            throw new Error(`Failed to update photo of menu item: ${id}.`);
        }

        alert(`Photo of menu item ${id} updated.`);
        loadMenu();
    } catch (error) {
        console.error(error);
        alert(`Error updating photo of menu item: ${id}.`);
    }
}


async function updateMenu(id){
    const name=document.getElementById(`name-${id}`).value;
    const description=document.getElementById(`description-${id}`).value;
    const price = parseFloat(document.getElementById(`price-${id}`).value);

    const ingText=document.getElementById(`ingredients-${id}`)?.value;
    let ingredients=[];
    try{
        ingredients = JSON.parse(ingText);
        if (!Array.isArray(ingredients)) throw new Error("Ingredients should be an array.");
    } catch (error) {
        alert("Invalid format for ingredients. Must be JSON array.");
        return;
    }

    const menu={
        name:name,
        description:description,
        price:price,
        ingredients:ingredients
    };

    try {
        const response = await fetch(`/menu/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(menu),
        });

        if (!response.ok) {
            throw new Error(`Failed to update menu item: ${id}.`);
        }

        alert("Menu item updated successfully!");
        loadMenu();
    } catch (error) {
        console.error(error);
        alert(`Error: Failed to update menu item: ${id}.`);
    }
}

async function deleteMenu(id) {
    if (confirm("Are you sure you want to delete the menu item?")) {
        try {
            const response = await fetch(`/menu/${id}`, { method: 'DELETE' });

            if (!response.ok) {
                throw new Error(`Failed to delete menu: ${id}.`);
            }

            alert(`Menu item ${id} deleted.`);
            loadMenu();
        } catch (error) {
            console.error(error);
            alert(`Error deleting menu item: ${id}.`);
        }
    }
}

async function removePhoto(id){
    if (confirm("Are you sure you want to delete the photo of menu item?")) {
        try {
            const response = await fetch(`/menu/${id}/image`, { method: 'DELETE' });

            if (!response.ok) {
                throw new Error(`Failed to delete Photo of menu: ${id}.`);
            }

            alert(`Photo of menu item ${id} deleted.`);
            loadMenu();
        } catch (error) {
            console.error(error);
            alert(`Error deleting Photo menu item: ${id}.`);
        }
    }
}