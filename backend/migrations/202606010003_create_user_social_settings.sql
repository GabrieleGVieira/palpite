create table if not exists user_social_settings (
	user_id uuid primary key,
	is_public_profile boolean not null default true,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);
