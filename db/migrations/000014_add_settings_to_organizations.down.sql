-- Migration: 000015_add_settings_to_organizations.down.sql
-- Purpose: Revert adding settings JSON column to organizations table

ALTER TABLE organizations DROP COLUMN settings;
