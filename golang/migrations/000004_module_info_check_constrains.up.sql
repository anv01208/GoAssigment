ALTER TABLE module_info
ADD CONSTRAINT check_updated_at_after_created_at
CHECK (updated_at >= created_at);
