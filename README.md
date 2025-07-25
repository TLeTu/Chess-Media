# Chess-Media

Chess-Media is a full-stack web application that allows users to play chess against each other in real-time or challenge a computer-controlled opponent. The project is built with a Go backend and a vanilla JavaScript frontend.

## Features

*   **Real-time Multiplayer:** Challenge other players to a game of chess.
*   **Ranked Matchmaking:** Enter the ranked queue to be automatically paired with an opponent of a similar skill level based on your ELO rating.
*   **Unranked Lobbies:** Create or join custom game rooms to play against friends.
*   **Play Against the Bot:** Hone your skills by playing against an AI opponent.
*   **User Authentication:** Secure user registration and login system.
*   **ELO Rating System:** Your ELO rating adjusts based on the outcome of your ranked matches.

## Tech Stack

### Backend
*   **Language:** Go
*   **Framework:** Gin
*   **Real-time Communication:** Gorilla WebSocket
*   **Database:** MySQL
*   **ORM:** GORM
*   **Authentication:** JWT (JSON Web Tokens)

### Frontend
*   **Core:** HTML5, CSS3, JavaScript (ES6+)
*   **Libraries:**
    *   Chessboard.js (for the interactive chessboard)
    *   jQuery (as a dependency for Chessboard.js)
    *   Bootstrap (for styling and layout)

## Project Structure

The project is divided into two main parts: the `client` and the `server`.

*   **`/client`**: Contains all the frontend assets, including HTML, CSS, and JavaScript files.
    *   `/client/pages`: HTML files for different views (home, game, login).
    *   `/client/src`: JavaScript files for handling game logic, UI interactions, and communication with the server.
    *   `/client/assets`: Static assets like CSS stylesheets and images.
*   **`/server`**: Contains the Go backend application.
    *   `main.go`: The entry point for the server application.
    *   `/ws`: Handles all WebSocket connections and real-time communication for game rooms.
    *   `/engine`: The core chess engine, responsible for game logic, move validation, and FEN handling.
    *   `/authentication`: Manages user registration, login, and JWT generation/validation.
    *   `/database`: Handles the database connection (MySQL) and data models using GORM.
    *   `/bot`: Contains the logic for the AI opponent.

## How to Run

1.  **Prerequisites:**
    *   Go (version 1.20 or later)
    *   MySQL Server

2.  **Backend Setup:**
    *   Navigate to the `server` directory.
    *   Create a `.env` file and configure your database connection details (DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME).
    *   Run `go mod tidy` to install dependencies.
    *   Run `go run main.go` to start the server. The server will run on `http://localhost:8080`.

3.  **Frontend Setup:**
    *   The frontend is composed of static files. Simply open the `index.html` file in the `client/pages` directory in your web browser, or navigate to `http://localhost:8080` after starting the backend server.