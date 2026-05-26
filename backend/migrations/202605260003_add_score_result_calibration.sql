alter table match_goal_predictions
	add column if not exists result_model_id uuid null references ml_models(id);

alter table match_goal_predictions
	add column if not exists result_probabilities jsonb null;

alter table match_goal_predictions
	add column if not exists calibration_method text null;

alter table match_goal_predictions
	add column if not exists score_probability_mass numeric null;

alter table match_goal_predictions
	add column if not exists calibrated_at timestamp null;

create index if not exists match_goal_predictions_result_model_idx
	on match_goal_predictions (result_model_id)
	where result_model_id is not null;

