drop table instance_settings;

delete from attachments where data is null;
alter table attachments drop constraint attachments_content;
alter table attachments drop column path;
alter table attachments alter column size type integer;
alter table attachments alter column data set not null;
