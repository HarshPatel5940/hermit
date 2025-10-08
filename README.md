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

This project is managed entirely through a comprehensive `Makefile`. Run `make help` to see all available commands.

### Prerequisites

*   Go (latest version)
*   Docker or Podman
*   `make`

### Installation & Running

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/harshpatel5940/hermit.git
    cd hermit
    ```

2.  **Set up environment variables:**
    Copy the `.env.example` file to `.env`. The default values are configured to work with the `docker-compose.yml` file and do not need to be changed for local development.

3.  **First-Time Setup:**
    This command starts all backend services (Postgres, MinIO, etc.) and runs the database migrations.
    ```sh
    make setup
    ```

4.  **Run the application in Development Mode:**
    This is the main command for development. It starts all services and runs the application with live-reloading. It will automatically rebuild the app and regenerate API documentation when you save a Go file.
    ```sh
    make dev
    ```

5.  **Access the API:**
    *   The API will be running at `http://localhost:8080`.
    *   API documentation is available at `http://localhost:8080/swagger/index.html`.

### Other Useful Commands

*   `make down`: Stop all running services.
*   `make logs`: Tail the logs from all running services.
*   `make test`: Run the test suite.
*   `make docs`: Manually regenerate the API documentation.
