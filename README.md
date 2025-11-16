# Social-With-Go

## Overview
Social-With-Go is a backend social networking application built using Go. It provides RESTful API endpoints with features such as user authentication, data management, and notification integration.

## Features
- Robust API development using Go and Chi router
- PostgreSQL database integration for persistent storage
- Automated migration and validation support
- Swagger-based API documentation
- Email notifications via SendGrid
- Logging with Uber Zap
- Rate limiting middleware implemented for API request control
- Containerized using Docker for easy deployment and environment consistency
- Unit tests and static analysis incorporated

## CI/CD
This project uses GitHub Actions for continuous integration and delivery. The workflow:
- Runs on every push and pull request to the master branch
- Verifies dependencies with `go mod verify`
- Builds the project using `go build`
- Runs static code analysis with `go vet` and `staticcheck`
- Executes tests with race condition detection

This ensures code quality and stability through automated testing and auditing before merging or deployment.

## Setup
1. Clone the repository  
2. Run `go mod download` to install dependencies  
3. Set up PostgreSQL and configure the connection  
4. Build and run the Docker container for the application  
5. Run migrations and start the server: `go run main.go` or `go build && ./Social-With-Go`

## Testing and Quality
- Static code analysis with `go vet` and `staticcheck`  
- Unit tests with race condition detection: `go test -race ./...`

## Contribution
Contributions are welcome. Please open issues or pull requests for improvements or fixes.

## License
This project is open source and available under the MIT License.

---
