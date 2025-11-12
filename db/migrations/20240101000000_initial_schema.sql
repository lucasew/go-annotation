-- migrate:up

-- Images table stores information about images to be annotated
CREATE TABLE images (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT UNIQUE NOT NULL,
  original_filename TEXT,
  ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  completed_stages INTEGER DEFAULT 0,
  is_finished BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_images_is_finished ON images(is_finished);
CREATE INDEX idx_images_completed_stages ON images(completed_stages);

-- Annotations table stores all annotations
-- Uses username directly from YAML config (no FK to users table)
CREATE TABLE annotations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  image_id INTEGER NOT NULL,
  username TEXT NOT NULL,
  stage_index INTEGER NOT NULL,
  option_value TEXT NOT NULL,
  annotated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(image_id, username, stage_index),
  FOREIGN KEY(image_id) REFERENCES images(id) ON DELETE CASCADE
);

CREATE INDEX idx_annotations_image_id ON annotations(image_id);
CREATE INDEX idx_annotations_username ON annotations(username);
CREATE INDEX idx_annotations_stage ON annotations(stage_index);

-- migrate:down

DROP INDEX IF EXISTS idx_annotations_stage;
DROP INDEX IF EXISTS idx_annotations_username;
DROP INDEX IF EXISTS idx_annotations_image_id;
DROP TABLE IF EXISTS annotations;

DROP INDEX IF EXISTS idx_images_completed_stages;
DROP INDEX IF EXISTS idx_images_is_finished;
DROP TABLE IF EXISTS images;
