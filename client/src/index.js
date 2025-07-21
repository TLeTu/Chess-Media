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
            method: 'GET',            headers: {
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

// --- Multiplayer Buttons ---

hostBtn.addEventListener("click", async function() {
    if (await validate()) {
        try {
            const response = await fetch('/api/rooms/create', { method: 'POST' });
            if (!response.ok) {
                throw new Error('Failed to create room');
            }
            const data = await response.json();
            window.location.href = `/game?room=${data.roomID}`;
        } catch (error) {
            console.error('Error creating room:', error);
            alert('Could not create a new room. Please try again.');
        }
    } else {
        window.location.href = '/login';
    }
});

joinBtn.addEventListener("click", async function() {
    if (await validate()) {
        const roomID = prompt("Please enter the Room ID to join:");
        if (roomID) {
            window.location.href = `/game?room=${roomID}`;
        }
    } else {
        window.location.href = '/login';
    }
});


// --- Placeholder for Ranked Button ---
rankBtn.addEventListener("click", async function() {
    if (await validate()) {
        alert("Ranked games are not yet implemented.");
    } else {
        window.location.href = '/login';
    }
});