create extension if not exists pgcrypto;

create table if not exists goal_models (
	id uuid primary key default gen_random_uuid(),
	name text not null,
	version text not null,
	algorithm text not null,
	artifact_path text not null,
	trained_from date null,
	trained_until date not null,
	feature_columns jsonb not null,
	metrics_json jsonb not null,
	status text not null default 'active',
	created_at timestamp default now()
);

create unique index if not exists goal_models_name_version_idx
	on goal_models (name, version);

create table if not exists match_goal_predictions (
	id uuid primary key default gen_random_uuid(),
	match_id uuid null,
	match_date date not null,
	home_team_id uuid not null references teams(id),
	away_team_id uuid not null references teams(id),
	goal_model_id uuid references goal_models(id),
	expected_home_goals numeric not null,
	expected_away_goals numeric not null,
	most_likely_home_score int null,
	most_likely_away_score int null,
	over_1_5_probability numeric null,
	over_2_5_probability numeric null,
	both_teams_score_probability numeric null,
	features_snapshot jsonb null,
	model_version text not null,
	source text not null default 'goals-ml-service',
	created_at timestamp default now(),
	updated_at timestamp default now()
);

create unique index if not exists match_goal_predictions_match_model_idx
	on match_goal_predictions (
		match_date,
		home_team_id,
		away_team_id,
		goal_model_id
	);

create table if not exists match_score_probabilities (
	id uuid primary key default gen_random_uuid(),
	match_goal_prediction_id uuid not null references match_goal_predictions(id) on delete cascade,
	home_score int not null,
	away_score int not null,
	probability numeric not null,
	created_at timestamp default now()
);

create unique index if not exists match_score_probabilities_score_idx
	on match_score_probabilities (
		match_goal_prediction_id,
		home_score,
		away_score
	);

