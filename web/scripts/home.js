  fetch('http://localhost:8080/reports/popular-items')
    .then(response => response.json())
    .then(data => {
      const popularItems = data.popular_items.slice(0, 10);
      const menuItemsContainer = document.getElementById("menu-carousel-inner");
      menuItemsContainer.innerHTML = "";

      for (let i = 0; i < popularItems.length; i += 5) {
        const itemsGroup = popularItems.slice(i, i + 5);
        const carouselItem = document.createElement('div');
        carouselItem.classList.add('carousel-item');
        if (i === 0) carouselItem.classList.add('active');

        const row = document.createElement('div');
        row.classList.add('d-flex', 'justify-content-center', 'gap-4');

        itemsGroup.forEach(item => {
          const card = document.createElement('div');
          card.classList.add('card', 'p-3');
          card.style.width = '22rem';

          const imagePath = item.image.startsWith("uploads/") ? `/${item.image}` : `/uploads/${item.image}`;
          const img = document.createElement('img');
          img.classList.add('card-img-top');
          img.src = imagePath;
          img.alt = item.name;
          img.style.height = '200px';
          img.style.objectFit = 'cover';

          const cardBody = document.createElement('div');
          cardBody.classList.add('card-body');

          const title = document.createElement('h5');
          title.classList.add('card-title');
          title.textContent = item.name;

          const description = document.createElement('p');
          description.classList.add('card-text');
          description.textContent = item.description;

          const quantity = document.createElement('p');
          quantity.classList.add('card-text');
          quantity.textContent = `Quantity: ${item.quantity}`;

          cardBody.appendChild(title);
          cardBody.appendChild(description);
          cardBody.appendChild(quantity);
          card.appendChild(img);
          card.appendChild(cardBody);
          row.appendChild(card);
        });

        carouselItem.appendChild(row);
        menuItemsContainer.appendChild(carouselItem);
      }
    })
    .catch(error => {
      console.error('Error fetching popular items:', error);
    });
