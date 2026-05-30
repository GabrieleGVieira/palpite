-- Run this in the Supabase SQL Editor for profile avatar uploads.
-- The app uploads files to: avatars/{auth.uid()}/avatar-{timestamp}.jpg

insert into storage.buckets (id, name, public)
values ('avatars', 'avatars', true)
on conflict (id) do update set public = excluded.public;

create policy "Authenticated users can upload own avatars"
on storage.objects
for insert
to authenticated
with check (
	bucket_id = 'avatars'
	and (storage.foldername(name))[1] = auth.uid()::text
);
