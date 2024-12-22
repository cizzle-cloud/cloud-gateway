# API Gateway

The **API Gateway** serves as the entrypoint for the infrastructure. Currently, it is implemented as a Go application that functions as a router with an integrated reverse proxy. The reverse proxy intercepts incoming HTTP requests and forwards them to the appropriate target host. It also allows for request modification, which currently includes modifications to the Host, Scheme, Header, and Path.

### Routes

The following routes are configured in the **API Gateway**:
- `/` redirects to `/access`
- `/access` forwards requests to `http://auth-page:80`
- `/auth/{PATH}` forwards requests to `http://auth:8000/{PATH}`


### Middleware

The **API Gateway** also supports middleware. Currently, two middleware methods are implemented to protected routes, **Auth Middleware** and **No-Cache Middleware**.
**Auth Middleware** intercepts requests to verify the client’s authentication status. The authentication cookie from the current session/context is retrieved. A new request is then created and sent to `/auth/validate` endpoint of **Auth service**, including the authentication cookie. If the response is successful, the request proceeds to the proxy handler. Otherwise the request is aborted.

**No-Cache Middleware** prevents the caching of sensitive data in protected endpoints by modifying HTTP headers. In this implementation, the middleware enforces a no-store, no-cache, must-revalidate policy, ensuring that browsers do not store data related to the session. The necessity of this middleware lies in the scenarios where a user has signed out or his session has expired, preventing unauthorized access to cached web applications on the client side.

### Future Improvements

- Single host reverse proxy is assumed as a good enough solution currently but maybe in the future these two concerns will be separated,  accommodating a more generic reverse proxy in front of the API Gateway.
- Rate Limiting and Throttling implementation. Rate limiting will control the number of requests a client can make within a specific time interval. Throttling will protect services from being overwhelmed by excessive traffic.
- Load Balancing in order to distribute incoming requests across multiple instances to enhance scalability.
- IP whitelisting/blacklisting
- API versioning
- Monitor and Analytics

### Installation

Clone the repository:

```bash
```

### Environment variables

The project requires the following environment variables. Create a `.env` file in the root directory and include:

```bash
PORT=<The port number on which the application will run inside the container>
FORWARDED_PORT=<The port on the host to which the container port is forwarded>
AUTH_ADDRESS=<The base URL for the authentication service>
AUTH_PAGE_ADDRESS=<The base URL for the authentication page>
API_GATEWAY_ADDRESS=<The base URL of the API Gateway that routes requests to microservices>
AUTH_PAGE_PATH=<The path to the authentication page>
AUTH_PATH=<The path to the authentication service>
```

### Running the application

The project uses `Dockerfile.dev` for the development environment, which sets up a containerized version of the app optimized for local development.

**Build and run the container**:

```bash
docker build -f Dockerfile.dev -t api-gateway-dev:<tag> .
docker run -v $(pwd):/api-gateway --env-file .env api-gateway-dev
```
The `-v` flag mounts the local root directory into the container, enabling changes to your local files to be reflected inside the container.
It is worth noting that API Gateway should run in conjuction with the rest of microservices.

**Stop the container**:
```
docker stop <container-id>
```

### Deployment
The production image created with `Dockerfile.prod` can be deployed to any container orchestration platform.

Ensure all required environment variables are set correctly in your orchestration platform.

```bash
docker build -f Dockerfile.prod -t api-gateway:<tag> .
```