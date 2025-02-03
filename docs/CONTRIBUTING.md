# Contributing to EventLoop  

- Install [Go](https://golang.org/doc/install) (ensure you use the version specified in `go.mod`).  
- Install [Docker](https://docs.docker.com/get-docker/) if working with the Docker setup.  

## Setup  

1. Clone the repository locally.  
2. Navigate to the project directory:  
    ```bash
    cd eventloop
    ```
3. Generate or provide a mock `.env` file with required variables.  
4. Cache Go modules:  
    ```bash
    go mod tidy
    ```
5. Build and run the project:  
    ```bash
    go run -v .
    ```

## Working on the Codebase  

- Create a new branch:  `git checkout -b yourname-feature`
- Add features and improvements—be descriptive, and include comments where necessary.  
- Commit your work in small, atomic chunks.  
- Follow the [Conventional Commit](https://gist.github.com/qoomon/5dfcdf8eec66a051ecd85625518cfd13#examples) format.  
- It's a good idea to pull changes to prevent merge conflicts
- Test changes before opening a PR.  
- Open the PR!