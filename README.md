```bash
go run main.go
curl http://localhost:8080/?addr=%EA%B2%BD%EA%B8%B0%EB%8F%84%20%EC%84%B1%EB%82%A8%EC%8B%9C%20%EB%B6%84%EB%8B%B9%EA%B5%AC%20%EC%84%9C%ED%98%84%EB%8F%99
docker build -t hongsgo/public_mask:latest .
docker run -p 8080:8080 hongsgo/public_mask:latest
```
