# Mangav5 (Wails v3)

**Mangav5** is a powerful manga downloader and manager built using [Wails v3](https://v3.wails.io/), combining the performance of Go with a modern Vue 3 frontend.

## 🚀 Features

*   **Customizable Scraping Engine**: Define your own scraping rules using JSON. Supports CSS selectors, Regex extraction, and template processing.
*   **Manga Management**: Track your favorite manga, chapters, and read status locally using SQLite.
*   **High-Performance Downloader**: Concurrent downloading of manga chapters and images.
*   **Modern UI**: Built with **Vue 3**, **TypeScript**, **Naive UI**, and **UnoCSS** for a responsive and clean user experience.
*   **Built-in Editor**: Integrated **Monaco Editor** for easy creation and modification of scraping rules directly within the app.
*   **Browser Integration**: Includes a browser service to handle complex sites and dynamic content.

## 🛠️ Tech Stack

*   **Backend**: Go (Golang), Wails v3
*   **Database**: SQLite (with auto-migrations)
*   **Frontend**: Vue 3, TypeScript, Vite
*   **UI Library**: Naive UI, UnoCSS
*   **Editor**: Monaco Editor

## 📦 Getting Started

### Prerequisites

*   **Go** (Latest version recommended)
*   **Node.js** & **npm/pnpm/yarn**
*   **Wails v3 CLI**

### Installation

1.  Clone the repository:
    ```bash
    git clone <repository-url>
    cd mangav5-wails3
    ```

2.  Install frontend dependencies:
    ```bash
    cd frontend
    npm install
    # or pnpm install / yarn install
    ```

3.  Run in Development Mode:
    ```bash
    wails3 dev
    ```
    This will start the application with hot-reloading enabled for both Go and Vue changes.

### Building for Production

To create a production-ready executable:

```bash
wails3 build
```
The output binary will be located in the `build/` directory.

## 📂 Project Structure

*   `main.go`: Application entry point and service registration.
*   `services/`: Core business logic (Scraper, Downloader, Database, Browser).
*   `internal/`: Internal packages for database models (`models/`) and repositories (`repo/`).
*   `frontend/`: Vue 3 application source code.
    *   `src/views/`: Application pages (Download, Library, etc.).
    *   `src/components/`: Reusable UI components.
*   `wails.json`: Wails project configuration.

## 📝 License

[Your License Here]
