-- Extensión requerida para generar UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Tabla de Usuarios (Custom Auth)
-- Implementamos nuestro propio modelo de usuario porque la bóveda Zero-Knowledge 
-- requiere guardar un AuthHash y el ClientSalt.
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    auth_hash VARCHAR(255) NOT NULL,
    client_salt VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Habilitar Row Level Security (RLS)
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- 2. Tabla de Entradas de la Bóveda
CREATE TABLE IF NOT EXISTS vault_entries (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_data TEXT NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Índices para optimizar las consultas de sincronización en Go
CREATE INDEX idx_vault_entries_user_id ON vault_entries(user_id);
CREATE INDEX idx_vault_entries_version ON vault_entries(version);
CREATE INDEX idx_vault_entries_updated_at ON vault_entries(updated_at);

-- Habilitar Row Level Security (RLS)
ALTER TABLE vault_entries ENABLE ROW LEVEL SECURITY;

-- ==========================================
-- POLÍTICAS DE SEGURIDAD (RLS)
-- Nota: Si usas Supabase Auth, se usa auth.uid(). 
-- Como estamos usando un backend Go que se conecta como admin (postgres),
-- el backend saltará el RLS, pero lo dejamos asegurado por buena práctica 
-- en caso de accesos directos desde la API de Supabase.
-- ==========================================

-- RLS para Users
CREATE POLICY "Users can view own record" ON users 
    FOR SELECT USING (auth.uid() = id);
    
CREATE POLICY "Users can update own record" ON users 
    FOR UPDATE USING (auth.uid() = id);

-- RLS para Vault Entries
CREATE POLICY "Users can view own entries" ON vault_entries 
    FOR SELECT USING (auth.uid() = user_id);
    
CREATE POLICY "Users can insert own entries" ON vault_entries 
    FOR INSERT WITH CHECK (auth.uid() = user_id);
    
CREATE POLICY "Users can update own entries" ON vault_entries 
    FOR UPDATE USING (auth.uid() = user_id);
    
CREATE POLICY "Users can delete own entries" ON vault_entries 
    FOR DELETE USING (auth.uid() = user_id);
