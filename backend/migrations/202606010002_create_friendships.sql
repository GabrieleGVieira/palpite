create table if not exists friendships (
	id uuid primary key default gen_random_uuid(),
	requester_user_id uuid not null,
	addressee_user_id uuid not null,
	status varchar not null check (status in ('PENDING', 'ACCEPTED', 'DECLINED', 'BLOCKED')),
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	constraint friendships_no_self_check check (requester_user_id <> addressee_user_id)
);

create unique index if not exists friendships_user_pair_unique_idx
	on friendships (
		least(requester_user_id, addressee_user_id),
		greatest(requester_user_id, addressee_user_id)
	);

create index if not exists friendships_requester_user_id_idx
	on friendships (requester_user_id);

create index if not exists friendships_addressee_user_id_idx
	on friendships (addressee_user_id);

create index if not exists friendships_status_idx
	on friendships (status);
