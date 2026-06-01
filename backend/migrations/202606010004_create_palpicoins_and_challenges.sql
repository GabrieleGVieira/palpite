create table if not exists user_wallets (
	id uuid primary key default gen_random_uuid(),
	user_id uuid not null unique,
	balance integer not null default 1000 check (balance >= 0),
	total_earned integer not null default 1000 check (total_earned >= 0),
	total_spent integer not null default 0 check (total_spent >= 0),
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table if not exists palpicoin_transactions (
	id uuid primary key default gen_random_uuid(),
	user_id uuid not null,
	amount integer not null check (amount <> 0),
	type varchar not null check (type in (
		'SIGNUP_BONUS',
		'MATCH_WINNER_HIT',
		'EXACT_SCORE_HIT',
		'ROUND_TOP_3',
		'CHALLENGE_STAKE',
		'CHALLENGE_WIN',
		'CHALLENGE_REFUND'
	)),
	description text not null default '',
	reference_type varchar,
	reference_id uuid,
	created_at timestamptz not null default now()
);

create index if not exists palpicoin_transactions_user_created_idx
	on palpicoin_transactions (user_id, created_at desc);

create unique index if not exists palpicoin_transactions_reference_unique_idx
	on palpicoin_transactions (user_id, type, reference_type, reference_id)
	where reference_type is not null and reference_id is not null;

insert into user_wallets (user_id, balance, total_earned)
select distinct gm.user_id, 1000, 1000
from group_members gm
where gm.status = 'active'
on conflict (user_id) do nothing;

insert into palpicoin_transactions (user_id, amount, type, description, reference_type, reference_id)
select uw.user_id, 1000, 'SIGNUP_BONUS', 'Bônus inicial de cadastro', 'user', uw.user_id
from user_wallets uw
on conflict (user_id, type, reference_type, reference_id)
	where reference_type is not null and reference_id is not null
do nothing;

create or replace function ensure_user_wallet(user_uuid uuid)
returns void
language plpgsql
as $$
begin
	insert into user_wallets (user_id, balance, total_earned)
	values (user_uuid, 1000, 1000)
	on conflict (user_id) do nothing;

	insert into palpicoin_transactions (user_id, amount, type, description, reference_type, reference_id)
	values (user_uuid, 1000, 'SIGNUP_BONUS', 'Bônus inicial de cadastro', 'user', user_uuid)
	on conflict (user_id, type, reference_type, reference_id)
		where reference_type is not null and reference_id is not null
	do nothing;
end $$;

create or replace function ensure_user_wallet_from_auth()
returns trigger
language plpgsql
as $$
begin
	perform ensure_user_wallet(new.id);
	return new;
end $$;

do $$
begin
	if exists (
		select 1
		from information_schema.tables
		where table_schema = 'auth'
			and table_name = 'users'
	) then
		drop trigger if exists ensure_user_wallet_after_auth_user_insert on auth.users;
		create trigger ensure_user_wallet_after_auth_user_insert
			after insert on auth.users
			for each row execute function ensure_user_wallet_from_auth();
	end if;
end $$;

create table if not exists palpicoin_challenges (
	id uuid primary key default gen_random_uuid(),
	creator_user_id uuid not null,
	opponent_user_id uuid not null,
	match_id uuid not null references world_cup_matches(id) on delete cascade,
	stake_amount integer not null check (stake_amount > 0),
	creator_prediction_id uuid,
	opponent_prediction_id uuid,
	creator_points integer,
	opponent_points integer,
	winner_user_id uuid,
	status varchar not null default 'PENDING' check (status in (
		'PENDING',
		'ACCEPTED',
		'DECLINED',
		'CANCELLED',
		'SETTLED'
	)),
	created_at timestamptz not null default now(),
	accepted_at timestamptz,
	settled_at timestamptz,
	updated_at timestamptz not null default now(),
	constraint palpicoin_challenges_no_self_check check (creator_user_id <> opponent_user_id)
);

create index if not exists palpicoin_challenges_creator_idx
	on palpicoin_challenges (creator_user_id, created_at desc);

create index if not exists palpicoin_challenges_opponent_idx
	on palpicoin_challenges (opponent_user_id, created_at desc);

create index if not exists palpicoin_challenges_match_status_idx
	on palpicoin_challenges (match_id, status);
