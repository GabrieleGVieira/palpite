create extension if not exists pgcrypto;

create table if not exists beta_testers_android (
	id uuid primary key default gen_random_uuid(),
	name text,
	email text not null unique,
	source text not null default 'landing',
	platform text not null default 'android',
	status text not null default 'pending_approval'
		check (status in ('pending', 'pending_approval', 'approved', 'rejected', 'added_to_google_group', 'exported', 'failed')),
	error_message text,
	approved_at timestamptz null,
	approved_by text null,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create index if not exists beta_testers_android_status_idx
	on beta_testers_android (status);
