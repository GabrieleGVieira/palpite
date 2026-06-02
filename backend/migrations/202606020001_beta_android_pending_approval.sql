alter table beta_testers_android
	add column if not exists platform text not null default 'android';

alter table beta_testers_android
	alter column status set default 'pending_approval';

alter table beta_testers_android
	drop constraint if exists beta_testers_android_status_check;

alter table beta_testers_android
	add constraint beta_testers_android_status_check
	check (status in ('pending', 'pending_approval', 'added_to_google_group', 'approved', 'exported', 'failed'));
