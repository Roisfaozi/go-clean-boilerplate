-- Multi-tenancy: Rollback Organizations and Members
-- Migration: 000011_create_organizations_tables.down.sql

DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
