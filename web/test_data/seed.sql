-- E2E Test Data Seed
-- Clear existing data
DELETE FROM sections;
DELETE FROM agency_cfr_references;
DELETE FROM agencies;

-- Parent agency (department)
INSERT INTO agencies VALUES ('test-department', 'Test Department', 'TD', 'Test Department', NULL);

-- Child agencies (sub-agencies)
INSERT INTO agencies VALUES ('test-bureau', 'Test Bureau', 'TB', 'Bureau, Test', 'test-department');
INSERT INTO agencies VALUES ('test-office', 'Test Office', 'TO', 'Office, Test', 'test-department');

-- CFR references linking agencies to titles/chapters
INSERT INTO agency_cfr_references VALUES ('test-department', 1, 'I');
INSERT INTO agency_cfr_references VALUES ('test-bureau', 1, 'II');
INSERT INTO agency_cfr_references VALUES ('test-office', 1, 'III');

-- Sections with metrics (join on title='1' and agency_id=chapter)
INSERT INTO sections VALUES (
  'test-section-1', '1', '100', '1.1', 'I', '/title-1/part-100/section-1.1',
  'Sample regulation text...', '2025-01-01', 'abc123', 5000, 10, 5, 3, 5180, 1036.0, '2025-01-01'
);
INSERT INTO sections VALUES (
  'test-section-2', '1', '200', '2.1', 'II', '/title-1/part-200/section-2.1',
  'Another regulation...', '2025-01-01', 'def456', 3000, 5, 2, 2, 3450, 1150.0, '2025-01-01'
);
INSERT INTO sections VALUES (
  'test-section-3', '1', '300', '3.1', 'III', '/title-1/part-300/section-3.1',
  'Third regulation...', '2025-01-01', 'ghi789', 2000, 3, 1, 1, 2400, 1200.0, '2025-01-01'
);
