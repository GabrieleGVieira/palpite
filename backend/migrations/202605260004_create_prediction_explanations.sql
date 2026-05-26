create extension if not exists pgcrypto;

create table if not exists prediction_explanations (
	id uuid primary key default gen_random_uuid(),
	match_id uuid null,
	match_prediction_id uuid null references match_predictions(id),
	goal_prediction_id uuid null references match_goal_predictions(id),
	home_team_id uuid not null references teams(id),
	away_team_id uuid not null references teams(id),
	match_date date not null,
	summary text not null,
	main_reasons jsonb not null,
	risk_alert text null,
	bet_style text null,
	user_tip text null,
	model_name text not null,
	prompt_version text not null,
	input_snapshot jsonb not null,
	raw_response jsonb null,
	status text not null default 'generated',
	error_message text null,
	created_at timestamp default now(),
	updated_at timestamp default now()
);

create unique index if not exists prediction_explanations_match_prompt_idx
	on prediction_explanations (
		match_date,
		home_team_id,
		away_team_id,
		prompt_version
	);

create index if not exists prediction_explanations_match_id_idx
	on prediction_explanations (match_id)
	where match_id is not null;

create index if not exists prediction_explanations_match_prediction_idx
	on prediction_explanations (match_prediction_id)
	where match_prediction_id is not null;

