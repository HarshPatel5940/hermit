# Hermit

> An intelligent, self-improving web crawler and archiver that powers a custom-trained chat AI.

---

## About The Project

Hermit is a powerful service designed to monitor, crawl, and archive websites. It systematically scrapes content from target sites, processes it into a structured format, and stores it for analysis and retrieval. The ultimate goal is to use this archived data to power a custom, fine-tuned large language model, enabling a sophisticated, context-aware chat and query interface for the scraped content.

This project is built with a focus on a robust, production-grade architecture from the ground up.

### Key Features

*   **Website Monitoring:** Add websites to a persistent monitoring list.
*   **Intelligent Crawling:** A `colly`-based crawler navigates and scrapes all pages of a target domain.
*   **Modern Data Pipeline:** Scraped content is stored in **MinIO**, with vector embeddings managed by **ChromaDB** for similarity search.
*   **Robust Backend:** A Go backend built with a clean, dependency-injected architecture using `uber-go/fx`.
*   **AI-Powered Chat (Future):** The stored data will be used to train and power a custom chat model using Retrieval-Augmented Generation (RAG).

## Technology Stack

*   **Backend:** Go
*   **Framework:** Echo
*   **Database:** PostgreSQL (managed with `sqlx`)
*   **Object Storage:** MinIO
*   **Vector Store:** ChromaDB
*   **API Documentation:** Swagger (`echo-swagger`)
*   **Live Reloading:** Air
*   **Containerization:** Docker & Docker Compose

## Getting Started

### Prerequisites

*   Go (latest version)
*   Docker and Docker Compose
*   `make`
*   `air` for live reloading (`go install github.com/air-verse/air@latest`)

### Installation & Running

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/harshnpatel/hermit.git
    cd hermit
    ```

2.  **Set up environment variables:**
    Copy the `.env.example` file to `.env` and fill in your database and service credentials. The default values are configured to work with the provided `docker-compose.yml` file.

3.  **Start backend services:**
    This will start PostgreSQL, MinIO, and ChromaDB in Docker containers.
    ```sh
    docker-compose up -d
    ```

4.  **Run database migrations:**
    This will create the necessary tables in your PostgreSQL database.
    ```sh
    go run cmd/migrate/main.go up
    ```

5.  **Run the application in development mode:**
    This command uses `air` for live reloading. It will automatically rebuild the app and regenerate API documentation when you save a Go file.
    ```sh
    air
    ```

6.  **Access the API:**
    *   The API will be running at `http://localhost:8080`.
    *   API documentation is available at `http://localhost:8080/swagger/index.html`.