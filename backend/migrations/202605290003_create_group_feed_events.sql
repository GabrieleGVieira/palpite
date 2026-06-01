create table if not exists group_feed_events (
	id uuid primary key default gen_random_uuid(),
	group_id uuid not null references groups(id) on delete cascade,
	event_type text not null check (event_type in (
		'member_joined',
		'leader_changed',
		'exact_score',
		'match_finished',
		'top3_reached'
	)),
	actor_user_id uuid,
	match_id uuid references world_cup_matches(id) on delete set null,
	metadata_json jsonb not null default '{}'::jsonb,
	created_at timestamptz not null default now()
);

create index if not exists group_feed_events_group_created_idx
	on group_feed_events (group_id, created_at desc);

create index if not exists group_feed_events_group_type_match_actor_idx
	on group_feed_events (group_id, event_type, match_id, actor_user_id);

create table if not exists group_feed_event_reactions (
	id uuid primary key default gen_random_uuid(),
	feed_event_id uuid not null references group_feed_events(id) on delete cascade,
	group_id uuid not null references groups(id) on delete cascade,
	user_id uuid not null,
	reaction_type text not null check (reaction_type in ('clap', 'fire', 'laugh', 'surprised', 'target')),
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	constraint group_feed_event_reactions_event_user_type_key
		unique (feed_event_id, user_id, reaction_type)
);

create index if not exists group_feed_event_reactions_event_idx
	on group_feed_event_reactions (feed_event_id);

create index if not exists group_feed_event_reactions_group_user_idx
	on group_feed_event_reactions (group_id, user_id);
