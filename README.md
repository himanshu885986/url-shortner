# URL Shortener (Go, In-Memory)

## Run

```bash
PORT=8080 BASE_URL=http://localhost:8080 go run ./cmd/server
```

## API

- POST `/api/v1/shorten`
  - body: `{ "url": "https://example.com/article" }`
  - resp: `{ "short_url": "http://localhost:8080/aB9", "code": "aB9" }`

- GET `/{code}`
  - 301 redirect to original URL

- GET `/api/v1/metrics`
  - gives the top requested Urls for shortening

## Notes
- In-memory store (not persistent). We can extend our application to use redis as storing mechanism
- Deterministic mapping: same long URL returns same code.
- Base62 codes from a monotonic counter.


## Dockerfile
- Use Below commands to run pull the image and run:
  - docker pull himanshu885986/url-shortener:latest
  - docker run -p 8080:8080 url-shortener