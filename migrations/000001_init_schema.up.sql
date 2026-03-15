CREATE TABLE users (
    id uuid PRIMARY KEY,
    google_sub text NOT NULL UNIQUE,
    email text NOT NULL,
    name text NOT NULL,
    avatar_url text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE photographer_profiles (
    user_id uuid PRIMARY KEY,
    username text NOT NULL UNIQUE,
    display_name text NOT NULL,
    bio text NOT NULL DEFAULT '',
    social_links jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_photographer_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    photographer_id uuid NOT NULL,
    title text NOT NULL,
    client_email text,
    status text NOT NULL DEFAULT 'draft',
    base_price_cents integer NOT NULL DEFAULT 0,
    included_count integer NOT NULL DEFAULT 0,
    extra_price_cents integer NOT NULL DEFAULT 0,
    min_select_count integer NOT NULL DEFAULT 0,
    currency text NOT NULL DEFAULT 'PLN',
    payment_mode text NOT NULL DEFAULT 'manual',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    closed_at timestamptz,
    delete_after timestamptz,

    CONSTRAINT fk_sessions_photographer
        FOREIGN KEY (photographer_id)
        REFERENCES users (id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_sessions_base_price_nonnegative
        CHECK (base_price_cents >= 0),

    CONSTRAINT chk_sessions_included_count_nonnegative
        CHECK (included_count >= 0),

    CONSTRAINT chk_sessions_extra_price_nonnegative
        CHECK (extra_price_cents >= 0),

    CONSTRAINT chk_sessions_min_select_count_nonnegative
        CHECK (min_select_count >= 0),

    CONSTRAINT chk_sessions_delete_after_requires_closed_at
        CHECK (delete_after IS NULL OR closed_at IS NOT NULL),

    CONSTRAINT chk_sessions_delete_after_after_closed_at
        CHECK (delete_after IS NULL OR closed_at IS NULL OR delete_after >= closed_at)
);

CREATE INDEX idx_sessions_photographer_created_at
    ON sessions (photographer_id, created_at DESC);

CREATE INDEX idx_sessions_delete_after
    ON sessions (delete_after);

CREATE TABLE session_access (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL,
    code_hmac text NOT NULL,
    token_hmac text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    last_used_at timestamptz,

    CONSTRAINT fk_session_access_session
        FOREIGN KEY (session_id)
        REFERENCES sessions (id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX ux_session_access_active_code_hmac
    ON session_access (code_hmac)
    WHERE revoked_at IS NULL;

CREATE UNIQUE INDEX ux_session_access_active_token_hmac
    ON session_access (token_hmac)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_session_access_session_created_at
    ON session_access (session_id, created_at DESC);

CREATE TABLE session_photos (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL,
    original_filename text NOT NULL,
    mime_type text NOT NULL,
    width integer,
    height integer,
    source_key text NOT NULL,
    source_size_bytes bigint NOT NULL,
    thumb_key text,
    proof_key text,
    status text NOT NULL DEFAULT 'uploaded',
    watermark_seed integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_session_photos_session
        FOREIGN KEY (session_id)
        REFERENCES sessions (id)
        ON DELETE CASCADE,

    CONSTRAINT uq_session_photos_session_id_id
        UNIQUE (session_id, id),

    CONSTRAINT chk_session_photos_width_positive
        CHECK (width IS NULL OR width > 0),

    CONSTRAINT chk_session_photos_height_positive
        CHECK (height IS NULL OR height > 0),

    CONSTRAINT chk_session_photos_source_size_nonnegative
        CHECK (source_size_bytes >= 0)
);

CREATE INDEX idx_session_photos_session_created_at
    ON session_photos (session_id, created_at DESC);

CREATE TABLE selections (
    session_id uuid NOT NULL,
    photo_id uuid NOT NULL,
    selected_at timestamptz NOT NULL DEFAULT now(),
    note text,

    CONSTRAINT pk_selections
        PRIMARY KEY (session_id, photo_id),

    CONSTRAINT fk_selections_photo_in_session
        FOREIGN KEY (session_id, photo_id)
        REFERENCES session_photos (session_id, id)
        ON DELETE CASCADE
);

CREATE INDEX idx_selections_session_selected_at
    ON selections (session_id, selected_at DESC);

CREATE TABLE payments (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL UNIQUE,
    method text NOT NULL DEFAULT 'manual',
    status text NOT NULL DEFAULT 'unpaid',
    amount_cents integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    paid_at timestamptz,

    CONSTRAINT fk_payments_session
        FOREIGN KEY (session_id)
        REFERENCES sessions (id)
        ON DELETE CASCADE,

    CONSTRAINT chk_payments_amount_nonnegative
        CHECK (amount_cents >= 0)
);

CREATE TABLE final_photos (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL,
    photo_id uuid NOT NULL,
    final_key text NOT NULL,
    final_size_bytes bigint,
    width integer,
    height integer,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT uq_final_photos_session_photo
        UNIQUE (session_id, photo_id),

    CONSTRAINT fk_final_photos_session
        FOREIGN KEY (session_id)
        REFERENCES sessions (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_final_photos_photo_in_session
        FOREIGN KEY (session_id, photo_id)
        REFERENCES session_photos (session_id, id)
        ON DELETE CASCADE,

    CONSTRAINT chk_final_photos_size_nonnegative
        CHECK (final_size_bytes IS NULL OR final_size_bytes >= 0),

    CONSTRAINT chk_final_photos_width_positive
        CHECK (width IS NULL OR width > 0),

    CONSTRAINT chk_final_photos_height_positive
        CHECK (height IS NULL OR height > 0)
);

CREATE INDEX idx_final_photos_session_created_at
    ON final_photos (session_id, created_at DESC);

CREATE TABLE deliveries (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL,
    version integer NOT NULL,
    status text NOT NULL DEFAULT 'generating',
    zip_key text,
    zip_size_bytes bigint,
    created_at timestamptz NOT NULL DEFAULT now(),
    generated_at timestamptz,

    CONSTRAINT fk_deliveries_session
        FOREIGN KEY (session_id)
        REFERENCES sessions (id)
        ON DELETE CASCADE,

    CONSTRAINT uq_deliveries_session_version
        UNIQUE (session_id, version),

    CONSTRAINT chk_deliveries_version_positive
        CHECK (version > 0),

    CONSTRAINT chk_deliveries_zip_size_nonnegative
        CHECK (zip_size_bytes IS NULL OR zip_size_bytes >= 0)
);

CREATE INDEX idx_deliveries_session_created_at
    ON deliveries (session_id, created_at DESC);

CREATE TABLE galleries (
    id uuid PRIMARY KEY,
    photographer_id uuid NOT NULL,
    title text NOT NULL,
    slug text NOT NULL,
    is_public boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_galleries_photographer
        FOREIGN KEY (photographer_id)
        REFERENCES users (id)
        ON DELETE RESTRICT,

    CONSTRAINT uq_galleries_photographer_slug
        UNIQUE (photographer_id, slug)
);

CREATE INDEX idx_galleries_photographer_created_at
    ON galleries (photographer_id, created_at DESC);

CREATE TABLE gallery_photos (
    id uuid PRIMARY KEY,
    gallery_id uuid NOT NULL,
    image_key text NOT NULL,
    width integer,
    height integer,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_gallery_photos_gallery
        FOREIGN KEY (gallery_id)
        REFERENCES galleries (id)
        ON DELETE CASCADE,

    CONSTRAINT chk_gallery_photos_width_positive
        CHECK (width IS NULL OR width > 0),

    CONSTRAINT chk_gallery_photos_height_positive
        CHECK (height IS NULL OR height > 0)
);

CREATE INDEX idx_gallery_photos_gallery_sort_order
    ON gallery_photos (gallery_id, sort_order, created_at);

CREATE TABLE jobs (
    id uuid PRIMARY KEY,
    type text NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    attempts integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 3,
    next_run_at timestamptz NOT NULL DEFAULT now(),
    locked_at timestamptz,
    locked_by text,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT chk_jobs_attempts_nonnegative
        CHECK (attempts >= 0),

    CONSTRAINT chk_jobs_max_attempts_positive
        CHECK (max_attempts > 0)
);

CREATE INDEX idx_jobs_dequeue
    ON jobs (status, next_run_at, created_at);

CREATE INDEX idx_jobs_locked_at
    ON jobs (locked_at)
    WHERE locked_at IS NOT NULL;
