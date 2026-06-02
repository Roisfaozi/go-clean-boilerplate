-- Multi-tenancy: Organizations and Members
-- Migration: 000011_create_organizations_tables.up.sql
-- Purpose: Create tables for multi-tenant architecture (Global User, Local Member model)

-- Core Organization Identity
CREATE TABLE organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    owner_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    INDEX idx_org_slug (slug),
    INDEX idx_org_owner (owner_id),
    INDEX idx_org_status (status)
);

-- Organization Members (Pivot Table)
-- Connects Global Users to specific Organizations with scoped roles
CREATE TABLE organization_members (
    id VARCHAR(36) PRIMARY KEY,
    organization_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    joined_at BIGINT,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_org_user (organization_id, user_id),
    INDEX idx_member_org (organization_id),
    INDEX idx_member_user (user_id),
    INDEX idx_member_status (status)
);
