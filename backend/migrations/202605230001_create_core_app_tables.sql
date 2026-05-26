create extension if not exists pgcrypto;

create table if not exists groups (
	id uuid primary key default gen_random_uuid(),
	owner_id uuid not null,
	name text not null,
	description text not null default '',
	match_scope text not null check (match_scope in ('all', 'selected')),
	selected_teams text[] not null default '{}',
	participant_limit integer check (participant_limit is null or participant_limit > 1),
	is_private boolean not null default true,
	invite_code text not null unique,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table if not exists group_members (
	group_id uuid not null references groups(id) on delete cascade,
	user_id uuid not null,
	role text not null check (role in ('owner', 'member')),
	display_name text not null default '',
	status text not null default 'active',
	joined_at timestamptz not null default now(),
	primary key (group_id, user_id)
);

create index if not exists group_members_user_status_idx
	on group_members (user_id, status);

create index if not exists group_members_group_status_idx
	on group_members (group_id, status);

create index if not exists group_members_group_user_status_idx
	on group_members (group_id, user_id, status);

create table if not exists world_cup_matches (
	id uuid primary key default gen_random_uuid(),
	external_id text,
	home_team text not null,
	away_team text not null,
	stage text not null,
	kickoff_at timestamptz not null,
	status text not null default 'scheduled'
		check (status in ('scheduled', 'live', 'finished', 'postponed', 'cancelled')),
	home_score integer check (home_score is null or (home_score >= 0 and home_score <= 99)),
	away_score integer check (away_score is null or (away_score >= 0 and away_score <= 99)),
	finished_at timestamptz,
	last_synced_at timestamptz,
	created_at timestamptz not null default now(),
	unique (home_team, away_team, kickoff_at)
);

create index if not exists world_cup_matches_status_kickoff_idx
	on world_cup_matches (status, kickoff_at);

create index if not exists world_cup_matches_kickoff_idx
	on world_cup_matches (kickoff_at);

create index if not exists world_cup_matches_external_id_idx
	on world_cup_matches (external_id)
	where external_id is not null;

create unique index if not exists world_cup_matches_external_id_unique_idx
	on world_cup_matches (external_id)
	where external_id is not null;

create table if not exists match_events (
	id uuid primary key default gen_random_uuid(),
	match_id uuid not null references world_cup_matches(id) on delete cascade,
	external_key text not null unique,
	event_type text not null check (event_type in ('goal', 'booking', 'substitution', 'penalty')),
	team_name text not null default '',
	player_name text not null default '',
	assist_name text not null default '',
	minute integer,
	injury_time integer,
	home_score integer,
	away_score integer,
	payload jsonb not null default '{}'::jsonb,
	created_at timestamptz not null default now()
);

create index if not exists match_events_match_id_idx
	on match_events (match_id, event_type, minute);

create table if not exists predictions (
	group_id uuid not null references groups(id) on delete cascade,
	match_id uuid not null references world_cup_matches(id) on delete cascade,
	user_id uuid not null,
	home_score integer not null check (home_score >= 0 and home_score <= 99),
	away_score integer not null check (away_score >= 0 and away_score <= 99),
	points integer,
	scored_at timestamptz,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	primary key (group_id, match_id, user_id)
);

create index if not exists predictions_user_group_idx
	on predictions (user_id, group_id);

create index if not exists predictions_group_user_idx
	on predictions (group_id, user_id);

create index if not exists predictions_match_idx
	on predictions (match_id);

create index if not exists predictions_group_match_idx
	on predictions (group_id, match_id);
