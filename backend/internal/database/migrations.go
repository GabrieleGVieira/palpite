package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		create extension if not exists pgcrypto;

		create table if not exists teams (
			id uuid primary key default gen_random_uuid(),
			name text not null unique,
			country_code text null,
			created_at timestamp default now(),
			updated_at timestamp default now()
		);

		create table if not exists team_aliases (
			id uuid primary key default gen_random_uuid(),
			team_id uuid not null references teams(id),
			alias text not null unique,
			created_at timestamp default now()
		);

		create table if not exists team_metrics (
			id uuid primary key default gen_random_uuid(),
			team_id uuid not null references teams(id),
			metric_date date not null,
			elo_score numeric null,
			attack_score numeric null,
			defense_score numeric null,
			recent_form_score numeric null,
			world_cup_history_score numeric null,
			knockout_score numeric null,
			group_stage_score numeric null,
			avg_goals_scored numeric null,
			avg_goals_conceded numeric null,
			win_rate numeric null,
			draw_rate numeric null,
			loss_rate numeric null,
			matches_played int default 0,
			source text not null default 'metrics-engine-v1',
			created_at timestamp default now(),
			updated_at timestamp default now()
		);

		create unique index if not exists team_metrics_team_date_idx
			on team_metrics (team_id, metric_date);

		create table if not exists team_metric_snapshots (
			id uuid primary key default gen_random_uuid(),
			team_id uuid not null references teams(id),
			snapshot_type text not null,
			payload_json jsonb not null,
			calculated_at timestamp default now()
		);

		create index if not exists team_metric_snapshots_team_type_calculated_idx
			on team_metric_snapshots (team_id, snapshot_type, calculated_at desc);

		create table if not exists match_features (
			id uuid primary key default gen_random_uuid(),
			match_id uuid null,
			match_date date not null,
			home_team_id uuid not null references teams(id),
			away_team_id uuid not null references teams(id),
			tournament text null,
			stage text null,
			home_elo_score numeric null,
			away_elo_score numeric null,
			elo_diff numeric null,
			home_attack_score numeric null,
			away_attack_score numeric null,
			home_defense_score numeric null,
			away_defense_score numeric null,
			home_recent_form_score numeric null,
			away_recent_form_score numeric null,
			home_fifa_rank int null,
			away_fifa_rank int null,
			fifa_rank_diff int null,
			home_avg_goals_scored numeric null,
			away_avg_goals_scored numeric null,
			home_avg_goals_conceded numeric null,
			away_avg_goals_conceded numeric null,
			home_world_cup_history_score numeric null,
			away_world_cup_history_score numeric null,
			neutral boolean default false,
			created_at timestamp default now(),
			updated_at timestamp default now()
		);

		create unique index if not exists match_features_match_unique_idx
			on match_features (
				match_date,
				home_team_id,
				away_team_id,
				tournament
			);

		create index if not exists match_features_match_date_idx
			on match_features (match_date);

		create table if not exists groups (
			id uuid primary key default gen_random_uuid(),
			owner_id uuid not null,
			name text not null,
			description text not null default '',
			match_scope text not null check (match_scope in ('all', 'selected')),
			selected_teams text[] not null default '{}',
			participant_limit integer check (participant_limit is null or participant_limit > 1),
			is_private boolean not null default true,
			is_paid boolean not null default false,
			payment_amount numeric(12,2) not null default 0 check (payment_amount >= 0),
			block_pending_predictions boolean not null default false,
			invite_code text not null unique,
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now()
		);

		alter table groups
			add column if not exists is_paid boolean not null default false;

		alter table groups
			add column if not exists payment_amount numeric(12,2) not null default 0
			check (payment_amount >= 0);

		alter table groups
			add column if not exists block_pending_predictions boolean not null default false;

		create table if not exists group_members (
			group_id uuid not null references groups(id) on delete cascade,
			user_id uuid not null,
			role text not null check (role in ('owner', 'member')),
			display_name text not null default '',
			status text not null default 'active',
			joined_at timestamptz not null default now(),
			primary key (group_id, user_id)
		);

		alter table group_members
			add column if not exists display_name text not null default '';

		alter table group_members
			add column if not exists status text not null default 'active';

		update group_members set status = 'active' where status = '';

		create index if not exists group_members_user_status_idx
			on group_members (user_id, status);

		create index if not exists group_members_group_status_idx
			on group_members (group_id, status);

		create index if not exists group_members_group_user_status_idx
			on group_members (group_id, user_id, status);

		create table if not exists group_payments (
			id uuid primary key default gen_random_uuid(),
			group_id uuid not null references groups(id) on delete cascade,
			user_id uuid not null,
			status text not null default 'pending' check (status in ('pending', 'paid', 'exempt', 'refunded')),
			amount_expected numeric(12,2) not null default 0 check (amount_expected >= 0),
			amount_paid numeric(12,2) not null default 0 check (amount_paid >= 0),
			payment_method text not null default '',
			paid_at timestamptz,
			marked_by_admin_id uuid,
			notes text not null default '',
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now(),
			unique (group_id, user_id),
			foreign key (group_id, user_id) references group_members(group_id, user_id) on delete cascade
		);

		create index if not exists group_payments_group_status_idx
			on group_payments (group_id, status);

		create index if not exists group_payments_group_user_idx
			on group_payments (group_id, user_id);

		create table if not exists world_cup_matches (
			id uuid primary key default gen_random_uuid(),
			external_id text,
			home_team text not null,
			away_team text not null,
			stage text not null,
			kickoff_at timestamptz not null,
			status text not null default 'scheduled' check (status in ('scheduled', 'live', 'finished', 'postponed', 'cancelled')),
			home_score integer check (home_score is null or (home_score >= 0 and home_score <= 99)),
			away_score integer check (away_score is null or (away_score >= 0 and away_score <= 99)),
			finished_at timestamptz,
			last_synced_at timestamptz,
			created_at timestamptz not null default now(),
			unique (home_team, away_team, kickoff_at)
		);

		alter table world_cup_matches
			add column if not exists external_id text;

		alter table world_cup_matches
			drop constraint if exists world_cup_matches_external_id_key;

		alter table world_cup_matches
			add column if not exists status text not null default 'scheduled'
			check (status in ('scheduled', 'live', 'finished', 'postponed', 'cancelled'));

		alter table world_cup_matches
			add column if not exists home_score integer check (home_score is null or (home_score >= 0 and home_score <= 99));

		alter table world_cup_matches
			add column if not exists away_score integer check (away_score is null or (away_score >= 0 and away_score <= 99));

		alter table world_cup_matches
			add column if not exists finished_at timestamptz;

		alter table world_cup_matches
			add column if not exists last_synced_at timestamptz;

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

		alter table predictions
			add column if not exists points integer;

		alter table predictions
			add column if not exists scored_at timestamptz;

		create index if not exists predictions_user_group_idx
			on predictions (user_id, group_id);

		create index if not exists predictions_group_user_idx
			on predictions (group_id, user_id);

		create index if not exists predictions_match_idx
			on predictions (match_id);

		create index if not exists predictions_group_match_idx
			on predictions (group_id, match_id);
	`)

	return err
}
