let botBtn = document.getElementById("botBtn");
let rankBtn = document.getElementById("rankBtn");
let joinBtn = document.getElementById("joinBtn");
let hostBtn = document.getElementById("hostBtn");

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

botBtn.addEventListener("click", function() {
    window.location.href = '/bot';
});

rankBtn.addEventListener("click", async function() {
    if(await validate()) {
        window.location.href = '/bot';
    }
    else {
        window.location.href = '/login';
    }
});

joinBtn.addEventListener("click", async function() {
    if(await validate()) {
        window.location.href = '/bot';
    }
    else {
        window.location.href = '/login';
    }
});

hostBtn.addEventListener("click", async function() {
    if(await validate()) {
        window.location.href = '/bot';
    }
    else {
        window.location.href = '/login';
    }
});