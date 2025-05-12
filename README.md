# Go_auth_service

# Как запустить в докере?
## Требования:
- Установленный docker
- Docker Desktop (для Windows)

## 1 Собираем контейнеры: 
```shell
docker-compose -f docker-compose.yml up -d
```
## 2 Чтобы посмотреть логи приложения:
```shell
docker-compose logs app
```

## Эндпоинты:
```POST /api/authorize/{userId} [авторизация не требуется]```
### response example: 200 OK
cookie:
```
refresh[readonly]: "129ueh9s8a7g8gsdAHVSDhsandkjni984rb..."
```
body:
```json
{
	"jwt":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzd..."
}
```


```POST /api/refresh_tokens [защищен]```

cookie (in request):
```
refresh[readonly]: "129ueh9s8a7g8gsdAHVSDhsandkjni984rb..."
```

request headers:
```
Authentification: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzd..."
```

### response example: 200 OK
cookie:
```
refresh[readonly]: "129ueh9s8a7g8gsdAHVSDhsandkjni984rb..."
```
body:
```json
{
	"jwt":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzd..."
}
```

```GET /api/guid [защищен]```

cookie (in request):
```
refresh[readonly]: "129ueh9s8a7g8gsdAHVSDhsandkjni984rb..."
```

request headers:
```
Authentification: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzd..."
```

### response example: 200 OK
```json
{
	"userId": "550e8400-e29b-41d4-a716-446655440000"
}
```


```POST /api/deauthorize [защищен]```

cookie (in request):
```
refresh[readonly]: "129ueh9s8a7g8gsdAHVSDhsandkjni984rb..."
```

request headers:
```
Authentification: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzd..."
```

### response example: 401 Unauthorized

```json
{
	"msg": "susseccfully deauthorized"
}
```