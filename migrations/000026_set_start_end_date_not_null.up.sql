UPDATE applications
SET project_start_date = '2024-01-01 00:00:00'
WHERE project_start_date IS NULL;
UPDATE applications
SET project_end_date = '2024-01-01 00:00:00'
WHERE project_end_date IS NULL;
alter table applications
modify project_start_date datetime not null default '2024-01-01 00:00:00';
alter table applications
modify project_end_date datetime not null default '2024-01-01 00:00:00';