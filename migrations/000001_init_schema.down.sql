DROP INDEX IF EXISTS idx_jobs_locked_at;
DROP INDEX IF EXISTS idx_jobs_dequeue;
DROP TABLE IF EXISTS jobs;

DROP INDEX IF EXISTS idx_gallery_photos_gallery_sort_order;
DROP TABLE IF EXISTS gallery_photos;

DROP INDEX IF EXISTS idx_galleries_photographer_created_at;
DROP TABLE IF EXISTS galleries;

DROP INDEX IF EXISTS idx_deliveries_session_created_at;
DROP TABLE IF EXISTS deliveries;

DROP INDEX IF EXISTS idx_final_photos_session_created_at;
DROP TABLE IF EXISTS final_photos;

DROP TABLE IF EXISTS payments;

DROP INDEX IF EXISTS idx_selections_session_selected_at;
DROP TABLE IF EXISTS selections;

DROP INDEX IF EXISTS idx_session_photos_session_created_at;
DROP TABLE IF EXISTS session_photos;

DROP INDEX IF EXISTS idx_session_access_session_created_at;
DROP INDEX IF EXISTS ux_session_access_active_token_hmac;
DROP INDEX IF EXISTS ux_session_access_active_code_hmac;
DROP TABLE IF EXISTS session_access;

DROP INDEX IF EXISTS idx_sessions_delete_after;
DROP INDEX IF EXISTS idx_sessions_photographer_created_at;
DROP TABLE IF EXISTS sessions;

DROP TABLE IF EXISTS photographer_profiles;
DROP TABLE IF EXISTS users;
