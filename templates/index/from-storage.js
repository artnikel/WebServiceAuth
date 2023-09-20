function testfunc() {
const accessToken = localStorage.getItem('access_token');

if (accessToken) {
    fetch('/index', {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${accessToken}`,
        },
    })
    .then(response => response.json())
    .then(data => {
        console.log(data);
    })
    .catch(error => {
        console.error(error);
    });
} else {
    console.error('No JWT token found in Local Storage');
}
};