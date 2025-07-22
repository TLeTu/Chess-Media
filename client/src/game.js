document.addEventListener('DOMContentLoaded', function() {
    let board = null;
    let socket = null;
    const statusEl = document.getElementById('status');
    const fenEl = document.getElementById('fen');
    let currentFen = 'start';
    let myColor = 'spectator'; // Default to spectator

    // --- WebSocket Connection ---
    function connect() {
        const urlParams = new URLSearchParams(window.location.search);
        const roomID = urlParams.get('room');

        if (!roomID) {
            alert('No room ID specified!');
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const wsURL = `${protocol}${window.location.host}/ws/game/${roomID}`;
        
        socket = new WebSocket(wsURL);

        socket.onopen = () => {
            console.log('WebSocket connection established');
            statusEl.textContent = 'Connected';
        };

        socket.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            handleServerMessage(msg);
        };

        socket.onclose = () => {
            console.log('WebSocket connection closed');
            statusEl.textContent = 'Disconnected';
        };

        socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            statusEl.textContent = 'Connection Error!';
        };
    }

    // --- Message Handling ---
    function handleServerMessage(message) {
        switch (message.action) {
            case 'player_assigned':
                myColor = message.payload.color;
                if (myColor === 'black') {
                    board.orientation('black');
                }
                document.querySelector('h1').textContent = `Multiplayer Chess (You are ${myColor})`;
                break;
            case 'game_state':
                updateGame(message.payload);
                break;
            case 'error':
                alert(`Error: ${message.payload.message}`);
                if (board) {
                    board.position(currentFen, false);
                }
                break;
            default:
                console.log('Unknown message action:', message.action);
        }
    }

    function sendMessage(action, payload) {
        if (socket && socket.readyState === WebSocket.OPEN) {
            const message = { action, payload };
            socket.send(JSON.stringify(message));
        }
    }

    // --- Game Logic and Board UI ---
    function onDragStart (source, piece, position, orientation) {
      // Do not pick up pieces if the game is over
      if (statusEl.textContent.includes('checkmate') || statusEl.textContent.includes('stalemate')) {
        return false
      }

      // Only pick up pieces for the side to move
      if ((position.turn === 'w' && piece.search(/^b/) !== -1) ||
          (position.turn === 'b' && piece.search(/^w/) !== -1)) {
        return false
      }
      
      // Only allow player to move their own pieces
      if ((myColor === 'white' && piece.search(/^b/) !== -1) ||
          (myColor === 'black' && piece.search(/^w/) !== -1)) {
          return false;
      }
    }

    function onDrop(source, target) {
        currentFen = board.fen();
        const move = { from: source, to: target };
        sendMessage('move', move);
    }

    function updateGame(gameState) {
        if (board) {
            board.position(gameState.fen, false);
        }
        currentFen = gameState.fen;
        statusEl.textContent = gameState.game_status.replace(/_/g, ' ');
        fenEl.textContent = gameState.fen;
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