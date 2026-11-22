ALTER TABLE students ADD COLUMN roll_number TEXT;

-- SQLite cannot add a UNIQUE column with ALTER TABLE, so uniqueness comes from an index.
-- Existing rows have NULL here, and a unique index allows many NULLs.
CREATE UNIQUE INDEX idx_students_roll_number ON students(roll_number);
