create extension if not exists pgcrypto;

create table if not exists ml_models (
	id uuid primary key default gen_random_uuid(),
	name text not null,
	version text not null,
	algorithm text not null,
	artifact_path text not null,
	trained_from date null,
	trained_until date not null,
	feature_columns jsonb not null,
	label_mapping jsonb not null,
	metrics_json jsonb not null,
	calibration_method text null,
	status text not null default 'active',
	created_at timestamp default now()
);

create unique index if not exists ml_models_name_version_idx
	on ml_models (name, version);

create table if not exists prediction_runs (
	id uuid primary key default gen_random_uuid(),
	model_id uuid not null references ml_models(id),
	status text not null,
	started_at timestamp default now(),
	finished_at timestamp null,
	matches_processed int default 0,
	error_message text null
);

create table if not exists historical_matches (
	id uuid primary key default gen_random_uuid(),
	match_date date not null,
	home_team_id uuid not null references teams(id),
	away_team_id uuid not null references teams(id),
	home_score int not null,
	away_score int not null,
	tournament text null,
	neutral boolean default false,
	source text not null default 'ml-service-local-csv',
	created_at timestamp default now(),
	updated_at timestamp default now()
);

create unique index if not exists historical_matches_match_unique_idx
	on historical_matches (
		match_date,
		home_team_id,
		away_team_id,
		coalesce(tournament, '')
	);

create table if not exists match_predictions (
	id uuid primary key default gen_random_uuid(),
	match_id uuid null,
	match_date date not null,
	home_team_id uuid not null references teams(id),
	away_team_id uuid not null references teams(id),
	model_id uuid references ml_models(id),
	home_win_probability numeric not null,
	draw_probability numeric not null,
	away_win_probability numeric not null,
	predicted_label text not null,
	confidence text not null,
	suggested_home_score int null,
	suggested_away_score int null,
	features_snapshot jsonb null,
	model_version text not null,
	source text not null default 'ml-service',
	created_at timestamp default now(),
	updated_at timestamp default now()
);

alter table match_predictions add column if not exists id uuid default gen_random_uuid();
alter table match_predictions add column if not exists match_id uuid null;
alter table match_predictions add column if not exists match_date date null;
alter table match_predictions add column if not exists home_team_id uuid null;
alter table match_predictions add column if not exists away_team_id uuid null;
alter table match_predictions add column if not exists model_id uuid null;
alter table match_predictions add column if not exists home_win_probability numeric null;
alter table match_predictions add column if not exists draw_probability numeric null;
alter table match_predictions add column if not exists away_win_probability numeric null;
alter table match_predictions add column if not exists predicted_label text null;
alter table match_predictions add column if not exists confidence text null;
alter table match_predictions add column if not exists suggested_home_score int null;
alter table match_predictions add column if not exists suggested_away_score int null;
alter table match_predictions add column if not exists features_snapshot jsonb null;
alter table match_predictions add column if not exists model_version text null;
alter table match_predictions add column if not exists source text default 'ml-service';
alter table match_predictions add column if not exists created_at timestamp default now();
alter table match_predictions add column if not exists updated_at timestamp default now();

do $$
begin
	if not exists (
		select 1 from pg_constraint
		where conname = 'match_predictions_model_id_fkey'
	) then
		alter table match_predictions
			add constraint match_predictions_model_id_fkey
			foreign key (model_id) references ml_models(id);
	end if;

	if not exists (
		select 1 from pg_constraint
		where conname = 'match_predictions_home_team_id_fkey'
	) then
		alter table match_predictions
			add constraint match_predictions_home_team_id_fkey
			foreign key (home_team_id) references teams(id);
	end if;

	if not exists (
		select 1 from pg_constraint
		where conname = 'match_predictions_away_team_id_fkey'
	) then
		alter table match_predictions
			add constraint match_predictions_away_team_id_fkey
			foreign key (away_team_id) references teams(id);
	end if;
end $$;

create unique index if not exists match_predictions_match_model_idx
	on match_predictions (
		match_date,
		home_team_id,
		away_team_id,
		model_id
	);
