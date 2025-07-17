let board = null;
let currentFen = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'; // Initialize with full starting FEN

const config = {
    draggable: true,
    position: 'start',
    onDrop: handlePlayerMove
};

document.addEventListener('DOMContentLoaded', function() {
    board = Chessboard('myBoard', config);
});

async function handlePlayerMove(source, target) {
    const playerMove = `${source}${target}`;
    const oldFen = currentFen; // Store the FEN before the move

    // Get the current board position object from the board instance
    const currentPositionObj = board.position();

    const sourceRank = parseInt(source[1]);
    const targetRank = parseInt(target[1]);
    const pieceOnSource = currentPositionObj[source]; // Get piece from the position object

    let promotionPiece = ''; // Default to no promotion

    // Check for pawn promotion
    if (pieceOnSource === 'wP' && sourceRank === 7 && targetRank === 8) {
        promotionPiece = promptForPromotion();
    } else if (pieceOnSource === 'bP' && sourceRank === 2 && targetRank === 1) {
        promotionPiece = promptForPromotion();
    }

    // If promotion was cancelled or invalid, snapback
    if (promotionPiece === null || (promotionPiece !== '' && !['q', 'r', 'b', 'n'].includes(promotionPiece.toLowerCase()))) {
        board.position(oldFen);
        return;
    }

    console.log('Sending FEN:', currentFen); // Debugging line

    const response = await fetch(`/api/bot/move`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            currentFen: currentFen,
            playerMove: playerMove,
            promotionPiece: promotionPiece // Send promotion choice
        })
    });

    if (!response.ok) {
        console.error("Invalid move!");
        board.position(oldFen);
        return;
    }

    const data = await response.json();
    handleBackendResponse(data);
}

function promptForPromotion() {
    let choice = prompt("Promote pawn to (Q, R, B, N)?", "Q");
    if (choice) {
        choice = choice.toLowerCase();
        if (['q', 'r', 'b', 'n'].includes(choice)) {
            return choice;
        }
    }
    return null; // User cancelled or entered invalid input
}

function handleBackendResponse(data) {
    currentFen = data.newFen; // newFen from backend should be a full FEN
    board.position(currentFen);

    if (data.gameStatus != 'in_progress') {
        alert(`Game Over: ${data.gameStatus}`)
    }
}