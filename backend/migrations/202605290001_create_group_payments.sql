alter table groups
	add column if not exists is_paid boolean not null default false;

alter table groups
	add column if not exists payment_amount numeric(12,2) not null default 0
	check (payment_amount >= 0);

alter table groups
	add column if not exists block_pending_predictions boolean not null default false;

create table if not exists group_payments (
	id uuid primary key default gen_random_uuid(),
	group_id uuid not null references groups(id) on delete cascade,
	user_id uuid not null,
	status text not null default 'pending'
		check (status in ('pending', 'paid', 'exempt', 'refunded')),
	amount_expected numeric(12,2) not null default 0
		check (amount_expected >= 0),
	amount_paid numeric(12,2) not null default 0
		check (amount_paid >= 0),
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
