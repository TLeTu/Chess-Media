let botBtn = document.getElementById("botBtn");
let rankBtn = document.getElementById("rankBtn");
let joinBtn = document.getElementById("joinBtn");
let hostBtn = document.getElementById("hostBtn");
let leaveQueueBtn = document.getElementById("leaveQueueBtn"); // New button

let queueStatusMessage = document.getElementById("queueStatusMessage");
let eloDisplay = document.getElementById("eloDisplay");
let eloValue = document.getElementById("eloValue");

// Function to update the queue status message
function updateQueueStatus(message) {
    if (queueStatusMessage) {
        queueStatusMessage.textContent = message;
    }
}

// Function to toggle button visibility
function toggleGameButtons(show) {
    rankBtn.classList.toggle('hidden', !show);
    hostBtn.classList.toggle('hidden', !show);
    joinBtn.classList.toggle('hidden', !show);
    botBtn.classList.toggle('hidden', !show);
    leaveQueueBtn.classList.toggle('hidden', show);
}

async function validate() {
    const token = localStorage.getItem('jwtToken');
    if (!token) {
        return { isValid: false };
    }
    try {
        const response = await fetch('/api/validate', {
            method: 'GET',            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        if (response.ok) {
            const data = await response.json();
            return { isValid: true, elo: data.elo };
        } else {
            return { isValid: false };
        }
    } catch (error) {
        console.error('Error during validation:', error);
        return { isValid: false };
    }
}

// Initial check for ELO display on page load
document.addEventListener('DOMContentLoaded', async () => {
    const validationResult = await validate();
    if (validationResult.isValid) {
        eloValue.textContent = validationResult.elo;
        eloDisplay.classList.remove('hidden');
    } else {
        eloDisplay.classList.add('hidden');
    }
});

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
let rankedWs = null; // Declare a variable to hold the WebSocket connection

rankBtn.addEventListener("click", async function() {
    const validationResult = await validate();
    if (!validationResult.isValid) {
        window.location.href = '/login';
        return;
    }

    // Clear any previous messages
    updateQueueStatus("Searching for a ranked opponent...");
    toggleGameButtons(false); // Hide game buttons, show leave queue button

    const token = localStorage.getItem('jwtToken');
    if (!token) {
        updateQueueStatus('You must be logged in to play ranked games.');
        toggleGameButtons(true); // Show game buttons again
        return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
    const wsURL = `${protocol}${window.location.host}/ws/game/ranked?token=${token}`; // Pass token as query parameter

    rankedWs = new WebSocket(wsURL); // Assign to the global variable

    rankedWs.onopen = () => {
        console.log('Connected to ranked queue WebSocket');
    };

    rankedWs.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        console.log('Received message from ranked queue:', msg);

        switch (msg.action) {
            case 'match_found':
                updateQueueStatus(`Match found! Redirecting...`);
                toggleGameButtons(true); // Show game buttons again
                // Redirect to the game page with the assigned room ID
                window.location.href = `/game?room=${msg.payload.roomID}`;
                break;
            case 'queue_status':
                // Display queue status messages to the user
                updateQueueStatus(`Queue Status: ${msg.payload.message}`);
                break;
            case 'error':
                updateQueueStatus(`Error: ${msg.payload.message}`);
                toggleGameButtons(true); // Show game buttons again
                break;
            default:
                console.log('Unknown message action:', msg.action);
        }
    };

    rankedWs.onclose = () => {
        console.log('Disconnected from ranked queue WebSocket');
        updateQueueStatus('Disconnected from queue.');
        toggleGameButtons(true); // Show game buttons again
    };

    rankedWs.onerror = (error) => {
        console.error('Ranked queue WebSocket error:', error);
        updateQueueStatus('An error occurred with the ranked queue. Please try again.');
        toggleGameButtons(true); // Show game buttons again
    };
});

leaveQueueBtn.addEventListener("click", function() {
    if (rankedWs && rankedWs.readyState === WebSocket.OPEN) {
        rankedWs.close(); // Close the WebSocket connection
        updateQueueStatus('Leaving ranked queue...');
    }
});