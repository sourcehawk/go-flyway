-- sql/afterMigrate.sql
-- Verify that the placeholder test inserted the correct, static values

DO $$
DECLARE
  got_txt       VARCHAR;
  row_count     INT;
  expected_list TEXT[] := ARRAY[
    'val1','val2','val3','val4',
    'val5','val6','val7','val8'
  ];
BEGIN
  -- 1) Check the single-row default/INSERT for id = 1
  SELECT txt
    INTO got_txt
  FROM ${flyway:defaultSchema}.variable_test
  WHERE id = 1;

  IF got_txt IS DISTINCT FROM 'test_value' THEN
    RAISE EXCEPTION
      '🛑 Placeholder test failed for id=1: expected "test_value" but got "%"', got_txt;
  END IF;

  -- 2) Check that exactly those eight values exist (regardless of id)
  SELECT COUNT(DISTINCT txt)
    INTO row_count
  FROM ${flyway:defaultSchema}.variable_test
  WHERE txt = ANY(expected_list);

  IF row_count <> array_length(expected_list, 1) THEN
    RAISE EXCEPTION
      '🛑 Multi-insert placeholder test failed: expected % distinct rows but found %',
      array_length(expected_list, 1), row_count;
  END IF;
END
$$;
