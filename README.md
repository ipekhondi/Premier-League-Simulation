# Premier League Simulation

A backend football league simulator written in Go, with Dockerized deployment and Postman-tested API endpoints.

## Features

- Simulates a Premier League season.
- Generates fixtures for the entire season.
- Provides API endpoints to initialize the league, play the next week, and get the league table.
- Persists data using a SQLite database.
- Dockerized for easy setup and deployment.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Go](https://golang.org/doc/install) (for running locally without Docker)
- A tool for making HTTP requests, like [Postman](https://www.postman.com/downloads/) or curl.

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/ipekhondi/premier-league-simulation.git
    cd premier-league-simulation
    ```

2.  **Using Docker (Recommended):**
    Build and run the container using Docker Compose:
    ```bash
    docker-compose up --build
    ```
    The application will be running on `http://localhost:8080`.

3.  **Running Locally (Without Docker):**
    - Install the Go dependencies:
      ```bash
      go mod tidy
      ```
    - Run the application:
      ```bash
      go run .
      ```
    The application will be running on `http://localhost:8080`.

## Usage

Once the application is running, you can interact with it using the API endpoints.

## API Endpoints

The following endpoints are available:

### Initialize League

- **URL:** `/api/league/initialize`
- **Method:** `POST`
- **Description:** Initializes the league with a set of teams and generates the fixtures for the season. This will create the initial data in the database.
- **Example Request (using curl):**
  ```bash
  curl -X POST http://localhost:8080/api/league/initialize
  ```

### Get League Table

- **URL:** `/api/league/table`
- **Method:** `GET`
- **Description:** Returns the current league table, sorted by points, goal difference, and goals for.
- **Example Request (using curl):**
  ```bash
  curl http://localhost:8080/api/league/table
  ```

### Play Next Week

- **URL:** `/api/league/play-next`
- **Method:** `POST`
- **Description:** Simulates the matches for the next week, updates the league table, and returns the results for the week.
- **Example Request (using curl):**
  ```bash
  curl -X POST http://localhost:8080/api/league/play-next
  ```

## Database

The application uses a SQLite database to store the league data. The database file (`league.db`) is located in the `data/` directory. The database is automatically created and migrated when the application starts.

The database schema consists of two tables:

- **`teams`**: Stores information about each team.
- **`matches`**: Stores information about each match, including the week, teams involved, and the result.

## Built With

- [Go](https://golang.org/) - The backend programming language.
- [GORM](https://gorm.io/) - The ORM library for Go.
- [SQLite](https://www.sqlite.org/index.html) - The database engine.
- [Docker](https://www.docker.com/) - The containerization platform.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
