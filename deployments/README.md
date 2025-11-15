Development:
- CD to project root and run the Go backend
- ```go run ./cmd/api/```
- CD to /web/app and run the Vite web server
- ```npm run dev```


Testing containerised app:
- Ensure Docker is running locally, eg Docker Deskop
- CD to the deployments directory
- ```docker compose up --build -d```