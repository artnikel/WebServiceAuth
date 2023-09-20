// document.addEventListener("DOMContentLoaded", function() {
//     if (window.location.pathname === '/login') {
//         test();
//     }
// });

function test() {
    console.log('begin')
    const login = document.getElementById('login').value;
    const password = document.getElementById('password').value;

    fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ login, password }),
    })
    .then(response => response.json())
    .then(data => {
        alert(JSON.stringify(data));
        const accessToken = data.access_token;
        localStorage.setItem('access_token', accessToken);
        window.location.href = '/index';
    })
    .catch(error => {
        console.error(error);
    });
};

// const loginBtn = document.getElementById('auth-button'); 

// loginBtn.addEventListener('submit', test);
