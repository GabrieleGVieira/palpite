alter table prediction_explanations
	add column if not exists ai_explanation_generated_at timestamp null,
	add column if not exists ai_explanation_version integer not null default 1,
	add column if not exists ai_explanation_retry_count integer not null default 0;

update prediction_explanations
set ai_explanation_generated_at = updated_at
where status = 'generated'
	and ai_explanation_generated_at is null;

create index if not exists prediction_explanations_refresh_idx
	on prediction_explanations (status, ai_explanation_generated_at);

create index if not exists prediction_explanations_retry_idx
	on prediction_explanations (status, ai_explanation_retry_count);
