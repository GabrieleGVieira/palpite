alter table beta_testers_android
	add column if not exists approved_at timestamptz null;

alter table beta_testers_android
	add column if not exists approved_by text null;

alter table beta_testers_android
	drop constraint if exists beta_testers_android_status_check;

alter table beta_testers_android
	add constraint beta_testers_android_status_check
	check (status in ('pending', 'pending_approval', 'approved', 'rejected', 'added_to_google_group', 'exported', 'failed'));
