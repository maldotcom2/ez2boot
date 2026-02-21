## Development
- CD to project root and run the Go backend 
    - ```go run ./cmd/api/```
- CD to /web/app and run the Vite web server 
    - ```npm run dev```

NOTE: To build/run on windows, requires C compiler on path

- Unzip to C:\Mingw64
https://github.com/niXman/mingw-builds-binaries/releases
https://github.com/niXman/mingw-builds-binaries/releases/download/15.2.0-rt_v13-rev0/x86_64-15.2.0-release-posix-seh-ucrt-rt_v13-rev0.7z

- Add C:\Mingw64\bin to path so you can run gcc from cli

## Dev testing locally against real cloud environment

#### Azure
- Ensure minimum environment variables are populated for Azure:
    - CLOUD_PROVIDER=azure
    - AZURE_SUBSCRIPTION_ID={your-subscription-from-azure-console}
- Install the Azure CLI if not already installed:
    - For Windows ```winget install Microsoft.AzureCLI```
- If CLI was freshly installed, close all CLIs including IDE and re-open.
- Login to Azure CLI 
    - ```az login```
- Run the Go backend and Vite web server as above, and ensure scraping is operational. LOG_LEVEL=debug to confirm.
- Don't forget to tag VMs.

## Dev Testing local containerised app:
- Ensure Docker is running locally, eg Docker Deskop.
- CD to the deployments directory and run the single command to build and bring the container online:
    - ```docker compose up --build -d```
- Alternatively, to run the build only, run the below command from the project root:
    - ```docker build -f deployments/Dockerfile -t ez2boot:test .```

NOTE: C compiler is not required for containerised testing as the build stage is performed within a container which already has one.