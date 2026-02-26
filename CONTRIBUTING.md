## Contributing

Contributions are welcome. This could be bug fixes or features, or simply raising awareness of a problem by creating an Issue.

### Structure
The backend is a Go application. Entry point is in cmd/api, with the majority of the backend code in internal. The backend follows a domain driven architecture, preferring receiver methods with DI where possible. The dev guide within the deployments directory contains info on getting started.
The frontend is a Vue application, in web/app and is also served by the Go app.

### Branches
This project uses a lightweight GitFlow strategy. The `main` branch represents the release branch and is always stable, `next` is the integration branch which PRs are merged into for CI testing and manual testing. Once `next` is stable and ready, it is merged back into `main` and a release is cut from `main`.

### Issues
Please try to be as descriptive as possible if the issue is for a bug, and include replication steps where possible.

### PRs
If you'd like to contribute code, please fork the repo, branch off `main` and raise a PR against `next`. Ideally, a PR will be related to an Issue but it isn't mandatory.
PRs should be small and targeted. I'm not considering any large sweeping changes and these will not be merged. If unsure, raise an Issue and let's chat.

### Rules
- Consistency over correctness. Follow the existing style.
- Readability over brevity or ceremony.
- No over-abstraction.

### AI
AI can be part of your workflow, but it's expected that all AI output has been fully read, understood, amended and tested before submission. If you wouldn't be able to reason about the code you're submitting, then don't submit it. If it looks like AI slop, it probably is and it won't be merged. 