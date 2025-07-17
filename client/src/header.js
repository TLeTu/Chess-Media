let logoutBtn = document.getElementById("logoutBtn")
let loginBtn = document.getElementById("loginBtn")

function logout() {
    localStorage.removeItem('jwtToken');
    window.location.href = '/login';
}

async function validate() {
    const token = localStorage.getItem('jwtToken');
    if (!token) {
        return false;
    }

    try {
        const response = await fetch('/api/validate', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        return response.ok;
    } catch (error) {
        console.error('Error during validation:', error);
        return false;
    }
}

logoutBtn.addEventListener('click', logout)

document.addEventListener('DOMContentLoaded', async () => {
    const loginBtn = document.querySelector('.login-btn');
    const logoutBtn = document.querySelector('.logout-btn');

    if (await validate()) {
        loginBtn.style.display = 'none';
        logoutBtn.style.display = 'block';
    } else {
        loginBtn.style.display = 'block';
        logoutBtn.style.display = 'none';
    }
});