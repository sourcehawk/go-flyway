-- V1__test_placeholder_replacements.sql
-- Flyway migration to test placeholder replacements
-- NOTE: The value of flyway:defaultSchema is set to the first value in -flyway.schemas
-- The migrator sets this argument according to the schema.name attribute in the config file
BEGIN;

-- 1. Create a test table in the configured schema
CREATE SCHEMA IF NOT EXISTS ${flyway:defaultSchema};
CREATE TABLE ${flyway:defaultSchema}.variable_test (
    id   SERIAL PRIMARY KEY,
    txt  VARCHAR(100) DEFAULT '${test_var}'
);

-- 2. Insert a row using the placeholder in a literal
INSERT INTO ${flyway:defaultSchema}.variable_test (txt)
VALUES ('${test_var}');

-- 3. Insert multiple rows using the placeholder in a literal
-- (Note: test_vars is a comma-separated list of values, e.g. (2, 'value1'), (3, 'value2'))
INSERT INTO ${flyway:defaultSchema}.variable_test (txt)
VALUES ${test_vars};

-- 3. Demonstrate using a placeholder in a comment/logging SELECT
--    (Not necessary for DDL, but shows that Flyway replaces everywhere)
SELECT
  '${test_var}'    AS replaced_value,
  current_schema() AS actual_schema
;

COMMIT;
