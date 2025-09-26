# Backend de NovelUzu

<div align="center">
  <p>Backend para plataforma de escritores y lectores de novelas web con sistema de publicación, descubrimiento y lectura en línea.</p>
  
  <p>
    <img src="https://img.shields.io/badge/Go-1.24-00ADD8?logo=go" alt="Go" />
    <img src="https://img.shields.io/badge/Gin-1.10-00ADD8?logo=go" alt="Gin" />
    <img src="https://img.shields.io/badge/PostgreSQL-16.9-336791?logo=postgresql" alt="PostgreSQL" />
    <img src="https://img.shields.io/badge/NextCloud-0082C9?logo=nextcloud" alt="NextCloud" />
    <img src="https://img.shields.io/badge/OpenNebula-6.10-C12D87?logo=opennebula" alt="OpenNebula" />
  </p>
  
  <p>
    <a href="#características">Características</a> •
    <a href="#despliegue">Despliegue</a> •
    <a href="#uso">Uso</a>
  </p>
</div>

## Dependencias principales

go version 1.24
gin-gonic (https://github.com/gin-gonic/gin) version 1.10
socket.io (github.com/zishang520/socket.io/v2) version 2.3.8

Las dependencias restantes se pueden encontrar en go.mod

#### Versiones externas relevantes

postgres version 16.9 (base de datos principal)
nextcloud (almacenamiento de imágenes)
servidor go alojado en OpenNebula 6.10

## Infraestructura

- **Base de datos**: PostgreSQL 16.9 para almacenamiento de datos de usuarios y aplicación
- **Almacenamiento de imágenes**: Nextcloud para gestión y almacenamiento de avatares y archivos multimedia
- **Servicio del sistema**: Configurado con systemd para reinicio automático y gestión del backend

## Despliegue

### Desarrollo local

Para desplegar este proyecto ejecuta:

```
go mod tidy
go run main.go

```
O compílalo como binario:
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

Generar documentación:
```
swag init --output config/swagger
```

## Características

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

## Uso/Ejemplos {#uso}

#### Documentación Swagger del servidor de desarrollo

~~~ copy

http://localhost:8080/swagger/index.html#/
~~~

#### Documentación Swagger del servidor de producción

~~~ copy

https://backnoveluzu.eslus.org/swagger/index.html#/
~~~
# Backend-NovelUzu
