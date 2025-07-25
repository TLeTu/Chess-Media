document.addEventListener('DOMContentLoaded', async () => {
    const loginBtn = document.getElementById('loginBtn');
    const logoutBtn = document.getElementById('logoutBtn');

    if (await validate()) {
        loginBtn.style.display = 'none';
        logoutBtn.style.display = 'block';
        logoutBtn.addEventListener('click', logout);
    } else {
        loginBtn.style.display = 'block';
        logoutBtn.style.display = 'none';
    }
});

async function validate() {
    try {
        const response = await fetch('/api/validate');
        return response.ok;
    } catch (error) {
        console.error('Error validating session:', error);
        return false;
    }
}

async function logout() {
    try {
        // Delete the JWT token from localStorage
        localStorage.removeItem('jwtToken');
        window.location.href = '/login';
    } catch (error) {
        console.error('Error during logout:', error);
    }
}