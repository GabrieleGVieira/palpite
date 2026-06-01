alter table group_feed_event_reactions
	drop constraint if exists group_feed_event_reactions_feed_event_id_user_id_key;

do $$
begin
	if not exists (
		select 1
		from pg_constraint
		where conname = 'group_feed_event_reactions_event_user_type_key'
	) then
		alter table group_feed_event_reactions
			add constraint group_feed_event_reactions_event_user_type_key
			unique (feed_event_id, user_id, reaction_type);
	end if;
end $$;
