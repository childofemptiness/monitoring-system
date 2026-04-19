ALTER TABLE monitors
ADD CONSTRAINT monitors_url_interval_unique
UNIQUE (url, interval_seconds);
