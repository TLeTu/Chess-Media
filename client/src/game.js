document.addEventListener('DOMContentLoaded', function() {
    let board = null;
    let socket = null;
    let myColor = 'spectator';
    let currentFen = 'start';

    // --- DOM Elements ---
    const lobbyContainer = document.getElementById('lobbyContainer');
    const gameContainer = document.getElementById('gameContainer');
    const lobbyStatus = document.getElementById('lobbyStatus');
    const hostControls = document.getElementById('hostControls');
    const startGameBtn = document.getElementById('startGameBtn');
    const readyBtn = document.getElementById('readyBtn');
    const statusEl = document.getElementById('status');
    const fenEl = document.getElementById('fen');
    const roomIDDisplay = document.getElementById('roomIDDisplay'); // New element

    // --- WebSocket Connection ---
    function connect() {
        const urlParams = new URLSearchParams(window.location.search);
        const roomID = urlParams.get('room');
        if (!roomID) {
            alert('No room ID specified!');
            return;
        }
        if (roomIDDisplay) {
            roomIDDisplay.textContent = roomID;
        }

        const token = localStorage.getItem('jwtToken');
        if (!token) {
            alert('Authentication token not found!');
            window.location.href = '/login';
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const wsURL = `${protocol}${window.location.host}/ws/game/${roomID}?token=${token}`;
        socket = new WebSocket(wsURL);

        socket.onopen = () => console.log('WebSocket connection established');
        socket.onclose = () => statusEl.textContent = 'Disconnected';
        socket.onerror = (error) => console.error('WebSocket error:', error);
        socket.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            handleServerMessage(msg);
        };
    }

    // --- Message Handling ---
    function handleServerMessage(message) {
        console.log('Received server message:', message.action, message.payload);
        switch (message.action) {
            case 'lobby_state':
                console.log('Updating lobby view. Game type:', message.payload.game_type);
                updateLobbyView(message.payload);
                break;
            case 'game_state':
                console.log('Updating game view. FEN:', message.payload.fen);
                updateGameView(message.payload);
                break;
            case 'error':
                // Error messages are now less intrusive
                console.warn('Server error:', message.payload.message);
                if (board && board.fen() !== currentFen) {
                    board.position(currentFen, false);
                }
                break;
            case 'player_assigned':
                myColor = message.payload.color;
                board.orientation(myColor === 'white' ? 'white' : 'black');
                // log to the console the color they got
                console.log('Assigned color:', myColor);
                break;
            default:
                console.log('Unknown message action:', message.action);
        }
    }

    function sendMessage(action, payload = {}) {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ action, payload }));
        }
    }

    // --- UI Update Functions ---
    function updateLobbyView(state) {
        // Only show lobby for non-ranked games
        if (state.game_type !== 'ranked') {
            lobbyContainer.classList.remove('hidden');
            gameContainer.classList.add('hidden');

            hostControls.classList.toggle('hidden', !state.is_host);
            readyBtn.classList.toggle('hidden', state.is_host);

            let statusText = `Players: ${state.player_count}/2.`;
            statusText += ` You are ${state.is_host ? 'the Host' : 'the Guest'}.`;
            lobbyStatus.innerHTML = `${statusText}<br>Guest is ${state.guest_ready ? 'Ready' : 'Not Ready'}.`;

            readyBtn.textContent = state.guest_ready ? 'Unready' : 'Ready';
        } else {
            // For ranked games, hide lobby and show game container immediately
            lobbyContainer.classList.add('hidden');
            gameContainer.classList.remove('hidden');
        }
    }

    function updateGameView(gameState) {
        lobbyContainer.classList.add('hidden');
        gameContainer.classList.remove('hidden');

        console.log('Board object:', board); // Check if board is initialized
        if (board) {
            board.position(gameState.fen, false);
            board.resize(); // Ensure the board redraws itself
        }
        currentFen = gameState.fen;
        const gameStatusText = gameState.game_status.replace(/_/g, ' ');
        statusEl.textContent = gameStatusText;
        fenEl.textContent = gameState.fen;

        if (gameState.game_status !== 'in_progress') {
            setTimeout(() => alert('Game Over: ' + gameStatusText), 300);
        }
    }

    // --- Event Listeners ---
    document.querySelectorAll('input[name="color"]').forEach(radio => {
        radio.addEventListener('change', (e) => sendMessage('assign_color', { color: e.target.value.toLowerCase() }));
    });
    readyBtn.addEventListener('click', () => sendMessage('player_ready'));
    startGameBtn.addEventListener('click', () => sendMessage('start_game'));

    // --- Game Logic and Board UI ---
    function onDragStart(source, piece, position, orientation) {
        if (myColor === 'spectator' || gameContainer.classList.contains('hidden')) return false;
        if (statusEl.textContent.includes('checkmate') || statusEl.textContent.includes('stalemate')) return false;
        if ((position.turn === 'w' && piece.search(/^b/) !== -1) || (position.turn === 'b' && piece.search(/^w/) !== -1)) return false;
        if ((myColor === 'white' && piece.search(/^b/) !== -1) || (myColor === 'black' && piece.search(/^w/) !== -1)) return false;
    }

    function onDrop(source, target) {
        currentFen = board.fen();
        const position = board.position();
        const piece = position[source];
        const sourceRank = source[1];
        const targetRank = target[1];
        const isPromotion = (piece === 'wP' && sourceRank === '7' && targetRank === '8') || (piece === 'bP' && sourceRank === '2' && targetRank === '1');

        let promotionChoice = '';
        if (isPromotion) {
            promotionChoice = prompt('Promote to? (q, r, b, n)', 'q');
            if (!promotionChoice || !['q', 'r', 'b', 'n'].includes(promotionChoice.toLowerCase())) {
                board.position(currentFen);
                return;
            }
        }
        sendMessage('move', { from: source, to: target, promotion: promotionChoice.toLowerCase() });
    }

    const config = {
        draggable: true,
        position: 'start',
        onDrop: onDrop,
        onDragStart: onDragStart,
    };

    board = Chessboard('myBoard', config);
    connect();
});