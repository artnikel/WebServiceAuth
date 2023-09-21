$(document).ready(function() {
    $("#updateBalance").click(function() {
      $("#balance").text(balance);
    });

    $("#deposit").click(function() {
      $("#depositModal").modal("show");
    });
  });

let userBalance = 0; 

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
    totalPrice += parseFloat(product.price) * product.quantity;
  }

  const totalPriceModal = document.getElementById('totalPriceModal');
  totalPriceModal.textContent = `Total Price: $${totalPrice.toFixed(2)}`;
}

document.getElementById('cartModal').addEventListener('click', () => {
  updateCartDisplay();
});

function calculateTotalAmount() {
  let total = 0;
  for (const productId in cart) {
    const product = cart[productId];
    total += product.price * product.quantity;
  }

  return total;
}

document.getElementById('buyButton').addEventListener('click', () => {
  const totalPrice = calculateTotalAmount(); 
  const data = { totalPrice }; 

  if (totalPrice === 0) {
    alert('Your cart is empty. Add items to the cart before making a purchase.');
    window.location.href = "/index"
    return;
  }

  fetch('/buy', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data), 
  })
    .then((response) => response.json())
    .then((data) => {
      if (data.success) {
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