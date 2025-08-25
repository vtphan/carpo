ALTER TABLE grades ADD COLUMN rating INTEGER DEFAULT 0 CHECK (rating >= -1 AND rating <= 1);;
