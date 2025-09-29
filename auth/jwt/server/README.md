POST /login
```shell
curl -i -X POST -H "Content-Type: application/json" \
  -d '{"username": "John"}' \
  localhost:8080/login
```

GET /protected
```shell
curl -i -H "Authorization: Bearer <token>" \
  localhost:8080/protected
```
