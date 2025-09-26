# NovelUzu's backend

## Core dependencies

go version 1.24
gin-gonic (https://github.com/gin-gonic/gin) version 1.10
socket.io (github.com/zishang520/socket.io/v2) version 2.3.8

The remaining dependencies can be found on go.mod

#### Relevant external versions

postgres version 16.9 (base de datos principal)
nextcloud (almacenamiento de imágenes)
go server hosted on OpenNebula 6.10

## Infrastructure

- **Base de datos**: PostgreSQL 16.9 para almacenamiento de datos de usuarios y aplicación
- **Almacenamiento de imágenes**: Nextcloud para gestión y almacenamiento de avatares y archivos multimedia
- **Servicio del sistema**: Configurado con systemd para reinicio automático y gestión del backend

## Deployment

### Desarrollo local

To deploy this project run:

```
go mod tidy
go run main.go

```
Or compile it as a binary:
```
go mod tidy
go build main.go
./main
```

### Servicio de producción

El backend está configurado como servicio systemd en `/etc/systemd/system/backend-noveluzu.service`:

```ini
[Unit]
Description=Backend NovelUzu
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/Backend-NovelUzu
ExecStart=/bin/bash -c "/usr/local/go/bin/go mod tidy && /usr/local/go/bin/go run main.go"
Restart=on-failure
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"

[Install]
WantedBy=multi-user.target
```

#### Comandos de gestión del servicio:

```bash
# Iniciar el servicio
sudo systemctl start backend-noveluzu

# Parar el servicio
sudo systemctl stop backend-noveluzu

# Reiniciar el servicio
sudo systemctl restart backend-noveluzu

# Ver estado del servicio
sudo systemctl status backend-noveluzu

# Habilitar arranque automático
sudo systemctl enable backend-noveluzu

# Ver logs del servicio
sudo journalctl -u backend-noveluzu -f
```

generate documentation:
```
swag init --output config/swagger
```

## Features

### Almacenamiento de imágenes con Nextcloud
- Integración completa con Nextcloud para subida y gestión de avatares
- Enlaces públicos automáticos para acceso a imágenes
- Validación de archivos (formato y tamaño)
- Gestión automática de directorios

### Base de datos PostgreSQL
- Almacenamiento seguro de datos de usuarios
- Gestión de sesiones y autenticación JWT
- Soporte completo para operaciones CRUD
- Integridad referencial y validaciones

### Servicio systemd
- Reinicio automático en caso de fallos
- Gestión centralizada del backend
- Logs estructurados con journalctl
- Arranque automático del sistema
## Usage/Examples

#### Development server swagger docs

~~~ copy

http://localhost:8080/swagger/index.html#/
~~~

#### Production server swagger docs

~~~ copy

https://backnoveluzu.eslus.org/swagger/index.html#/
~~~
# Backend-NovelUzu
