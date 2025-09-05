-- Base de datos NovelUzu - Esquema completo PostgreSQL
-- Creación de tipos enumerados
CREATE TYPE user_role AS ENUM ('usuario', 'admin');
CREATE TYPE user_status AS ENUM ('activo', 'inactivo', 'suspendido', 'baneado');
CREATE TYPE novel_status AS ENUM ('en_progreso', 'completada', 'pausada', 'abandonada');
CREATE TYPE chapter_status AS ENUM ('borrador', 'publicado', 'programado');
CREATE TYPE comment_status AS ENUM ('pendiente', 'aprobado', 'rechazado', 'oculto');
CREATE TYPE report_status AS ENUM ('pendiente', 'en_revision', 'resuelto', 'descartado');
CREATE TYPE report_type AS ENUM ('spam', 'inapropiado', 'copyright', 'acoso', 'otro');
CREATE TYPE notification_type AS ENUM ('nuevo_capitulo', 'respuesta_comentario', 'actualizacion_novela', 'sistema', 'logro');
CREATE TYPE subscription_status AS ENUM ('activa', 'cancelada', 'expirada');
-- Tabla de usuarios
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role DEFAULT 'usuario',
    status user_status DEFAULT 'activo',
    avatar_url TEXT,
    bio TEXT,
    birth_date DATE,
    country VARCHAR(100),
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
