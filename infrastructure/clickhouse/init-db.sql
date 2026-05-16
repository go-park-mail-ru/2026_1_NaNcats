CREATE DATABASE IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.schema_migrations (
    version Int64,
    dirty UInt8,
    sequence UInt64
) ENGINE = MergeTree() 
ORDER BY version;
