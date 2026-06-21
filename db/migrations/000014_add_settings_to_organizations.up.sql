-- Migration: 000015_add_settings_to_organizations.up.sql
-- Purpose: Add settings JSON column to organizations table

ALTER TABLE organizations ADD COLUMN settings JSON;
