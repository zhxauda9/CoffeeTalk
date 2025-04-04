async function fetchMenu() {
    try {
      const response = await fetch('http://localhost:8080/menu');
      const data = await response.json();
      const menuContainer = document.getElementById('menu');
      menuContainer.innerHTML = '';
      data.forEach(item => {
        const card = document.createElement('div');
        card.classList.add('menu-card');
        const imagePath = item.image.startsWith("uploads/") ? `/${item.image}` : `/uploads/${item.image}`;
        const img = document.createElement('img');
        img.src = imagePath;
        img.alt = item.name;
        const body = document.createElement('div');
        body.classList.add('card-body');
        const title = document.createElement('h5');
        title.textContent = item.name;
        const description = document.createElement('p');
        description.textContent = item.description;
        const price = document.createElement('p');
        price.textContent = `$${item.price}`;
        body.appendChild(title);
        body.appendChild(description);
        body.appendChild(price);
        card.appendChild(img);
        card.appendChild(body);
        menuContainer.appendChild(card);
      });
    } catch (error) {
      console.error('Error fetching menu:', error);
    }
  }
  
  fetchMenu();