let botBtn = document.getElementById("botBtn");
let rankBtn = document.getElementById("rankBtn");
let joinBtn = document.getElementById("joinBtn");
let hostBtn = document.getElementById("hostBtn");

let queueStatusMessage = document.getElementById("queueStatusMessage");

// Function to update the queue status message
function updateQueueStatus(message) {
    if (queueStatusMessage) {
        queueStatusMessage.textContent = message;
    }
}

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

botBtn.addEventListener("click", async function() {
    const token = localStorage.getItem('jwtToken');
    if (token) {
        window.location.href = '/bot';
    } else {
        window.location.href = '/login';
    }
});

// --- Multiplayer Buttons ---

hostBtn.addEventListener("click", async function() {
    const token = localStorage.getItem('jwtToken');
    if (token) {
        try {
            const response = await fetch('/api/rooms/create', { 
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });
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
    const token = localStorage.getItem('jwtToken');
    if (token) {
        const roomID = prompt("Please enter the Room ID to join:");
        if (roomID) {
            window.location.href = `/game?room=${roomID}`;
        }
    } else {
        window.location.href = '/login';
    }
});


// --- Ranked Button ---
rankBtn.addEventListener("click", async function() {
    if (!(await validate())) {
        window.location.href = '/login';
        return;
    }

    // Clear any previous messages
    updateQueueStatus("Searching for a ranked opponent...");

    const token = localStorage.getItem('jwtToken');
    if (!token) {
        updateQueueStatus('You must be logged in to play ranked games.');
        return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
    const wsURL = `${protocol}${window.location.host}/ws/game/ranked?token=${token}`; // Pass token as query parameter

    const ws = new WebSocket(wsURL);

    ws.onopen = () => {
        console.log('Connected to ranked queue WebSocket');
        // updateQueueStatus('Searching for a ranked opponent...'); // Already set above
        // Optionally disable the ranked button or show a loading indicator
    };

    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        console.log('Received message from ranked queue:', msg);

        switch (msg.action) {
            case 'match_found':
                updateQueueStatus(`Match found! Redirecting...`);
                // Redirect to the game page with the assigned room ID
                window.location.href = `/game?room=${msg.payload.roomID}`;
                break;
            case 'queue_status':
                // Display queue status messages to the user
                updateQueueStatus(`Queue Status: ${msg.payload.message}`);
                break;
            case 'error':
                updateQueueStatus(`Error: ${msg.payload.message}`);
                // Re-enable the ranked button or hide loading indicator
                break;
            default:
                console.log('Unknown message action:', msg.action);
        }
    };

    ws.onclose = () => {
        console.log('Disconnected from ranked queue WebSocket');
        updateQueueStatus('Disconnected from queue.');
        // Re-enable the ranked button or hide loading indicator
    };

    ws.onerror = (error) => {
        console.error('Ranked queue WebSocket error:', error);
        updateQueueStatus('An error occurred with the ranked queue. Please try again.');
        // Re-enable the ranked button or hide loading indicator
    };
});