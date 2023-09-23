$(document).ready(function() {
    $("#updateBalance").click(function() {
      $("#balance").text(balance);
    });

    $("#deposit").click(function() {
      $("#depositModal").modal("show");
    });
  });

let cart = {};

document.querySelectorAll('[data-product-id]').forEach(button => {
  button.addEventListener('click', (event) => {
    const productId = event.target.getAttribute('data-product-id');

    const productInfo = getProductInfo(productId);

    if (cart[productId]) {
      cart[productId].quantity++;
    } else {
      cart[productId] = {
        name: productInfo.name,
        price: parseFloat(productInfo.price.replace('$', '')), 
        image: productInfo.image,
        quantity: 1
      };
    }
    updateCartDisplay();
  });
});

function getProductInfo(productId) {
  if (productId === 'huawei-laptop') {
    return {
      name: 'HUAWEI MATEBOOK D 16',
      price: '$949.99',
      image: '/templates/index/images/huaweilaptop1.jpg'
    };
  } else if (productId === 'asus-laptop') {
    return {
      name: 'ASUS ROG STRIX',
      price: '$1099.99',
      image: '/templates/index/images/asuslaptop1.jpg'
    };
  } else if (productId === 'lenovo-laptop') {
    return {
      name: 'LENOVO IDEAPAD SLIM 3',
      price: '$499.99',
      image: '/templates/index/images/lenovolaptop1.jpeg'
    };
  } else if (productId === 'razer-mouse') {
    return {
      name: 'RAZER DEATHADDER ESSENTIAL',
      price: '$34.99',
      image: '/templates/index/images/razermouse1.jpg'
    };
  } else if (productId === 'logi-mouse') {
    return {
      name: 'LOGITECH G102 LIGHTSYNC',
      price: '$29.99',
     image: '/templates/index/images/logimouse1.jpg'
    };
  } else if (productId === 'bloody-mouse') {
    return {
      name: 'A4TECH BLOODY R90 PLUS',
      price: '$54.99',
     image: '/templates/index/images/bloodymouse1.jpg'
    };
  } else if (productId === 'asus-headphones') {
    return {
      name: 'ASUS TUF GAMING H3',
      price: '$74.99',
      image: '/templates/index/images/asusheadphones1.jpg'
    };
  } else if (productId === 'razer-headphones') {
    return {
      name: 'RAZER BLACKSHARK V2 PRO',
      price: '$169.99',
      image: '/templates/index/images/razerheadphones1.jpg'
    };
  } else if (productId === 'jbl-headphones') {
    return {
      name: 'JBL QUANTUM 100',
      price: '$49.99',
      image: '/templates/index/images/jblheadphones1.jpg'
    };
  }
}

function updateCartDisplay() {
  const cartItemList = document.getElementById('cartItemList');
  cartItemList.innerHTML = '';

  let totalPrice = 0;

  for (const productId in cart) {
    const product = cart[productId];
    const listItem = document.createElement('li');

    const productImage = document.createElement('img');
    productImage.src = product.image;
    productImage.width = 60;
    productImage.height = 50;

    const productDescription = document.createTextNode(`${product.name} - $${product.price} x ${product.quantity}`);

    listItem.appendChild(productImage);
    listItem.appendChild(productDescription);

    cartItemList.appendChild(listItem);
    totalPrice += product.price * product.quantity;
  }

  const totalPriceModal = document.getElementById('totalPriceModal');
  totalPriceModal.textContent = `Total Price: $${totalPrice.toFixed(2)}`;
}

document.getElementById('cartModal').addEventListener('click', () => {
  updateCartDisplay();
});

document.getElementById('buyButton').addEventListener('click', () => {
  const totalPriceElement = document.getElementById('totalPriceModal');
  const totalPriceText = totalPriceElement.textContent;
  const totalPrice = parseFloat(totalPriceText.match(/\d+\.\d+/)[0]); 

  if (isNaN(totalPrice) || totalPrice === 0) {
    alert('Your cart is empty or the total price is invalid. Add items to the cart before making a purchase.');
    return;
  }
  fetch('/buy', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ totalPrice }),
  })
    .then((response) => response.json())
    .then((data) => {
      if (data.success) {
        alert('Purchase successfully completed');
      } else {
        alert('Unable to complete the purchase. Not enough money');
      }
    })
    .catch((error) => {
      console.error('Error:', error);
    });
});


function clearCart() {
  cart = {}; 
  updateCartDisplay(); 
}

function buyWithPost() {
  const form = document.createElement('form');
  form.method = 'POST';
  form.action = '/buy';
  document.body.appendChild(form);
  form.submit();
}

document.getElementById('saveButton').addEventListener('click', () => {
  const cartItems = [];

  for (const productId in cart) {
    const product = cart[productId];
    const productPrice = typeof product.price === 'string' ? parseFloat(product.price.replace('$', '')) : product.price;
    cartItems.push({
      product_image: product.image, 
      product_name: product.name,
      product_price: productPrice,
      quantity: product.quantity
    });
  }

  fetch('/savecart', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(cartItems),
  })
  .then((response) => response.text())
  .then((data) => {
  })
  .catch((error) => {
    console.error('Error:', error);
  });
});

function refreshCart(){
  const cartLink = document.getElementById('cartLink');
  cartLink.click();
  window.location.href = "/index"
}


