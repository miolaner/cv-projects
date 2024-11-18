Literary Lions Forum
====================

Project Description
-------------------

**Literary Lions Forum** is a simple, discussion platform where users can create accounts, post topics, leave comments, and interact with others by liking or disliking posts. Users also have the ability to view their posts and the posts they've liked.

This forum was built using Go, SQLite, and HTML/CSS. It is containerized using Docker for easy deployment and comes with features such as authentication, session management, and moderator controls.

Features
--------

*   User authentication (registration, login, logout)
    
*   Create, view, and delete posts
    
*   Like and dislike posts
    
*   Comment on posts
    
*   User profiles with editable personal information
    
*   Moderator functionality to delete inappropriate comments and posts
    
*   Search functionality for finding posts
    
*   Dockerized for easy deployment
    

Tech Stack
----------

*   **Backend**: Go (Golang)
    
*   **Database**: SQLite
    
*   **Frontend**: HTML, CSS
    
*   **Containerization**: Docker
    

Prerequisites
-------------

To run this project locally, ensure you have the following installed:

*   Go (version 1.20+)
    
*   Docker
    
*   SQLite
    

Setup
-----

### 1\. Clone the repository

```sh
git clone https://gitea.koodsisu.fi/miolaner/literary-lions
cd literary-lions
```

### 2\. Build and Run with Docker

```sh
docker build -t literary-lions .
```
```sh
docker run -p 8080:8080 literary-lions
```
### 3\. Running locally without docker

1.  Install Go dependencies:
    
```sh
go mod download
```

1.  Run the application:
    
```sh
go run .
```

The application should be accessible at http://localhost:8080.

### 5\. Usage

*   Open your browser and navigate to http://localhost:8080.
    
*   Register for an account.
    
*   Create new posts and interact with others by liking, disliking, and commenting on posts.

*   You can login to testuser password:salasana to see moderator features like deleting comments and posts

UUID-Based Sessions
-------------------

The project uses UUID-based session management to securely handle user sessions. Each session is uniquely identified, and session cookies ensure that users remain logged in across different sessions.