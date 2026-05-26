from __future__ import annotations

import json
from datetime import date

from app.ml.database import Database, _json_safe


def load_prediction_pairs(
    db: Database,
    *,
    result_model_id: str,
    goal_model_id: str,
    from_date: date,
    to_date: date,
) -> list[dict]:
    with db.connect() as conn:
        rows = conn.execute(
            """
            select
                mgp.id::text as match_goal_prediction_id,
                mgp.match_id::text as match_id,
                mgp.match_date,
                mgp.home_team_id::text as home_team_id,
                mgp.away_team_id::text as away_team_id,
                mgp.goal_model_id::text as goal_model_id,
                mgp.expected_home_goals::float8 as expected_home_goals,
                mgp.expected_away_goals::float8 as expected_away_goals,
                mp.model_id::text as result_model_id,
                mp.home_win_probability::float8 as home_win_probability,
                mp.draw_probability::float8 as draw_probability,
                mp.away_win_probability::float8 as away_win_probability
            from match_goal_predictions mgp
            join match_predictions mp on mp.match_date = mgp.match_date
                and mp.home_team_id = mgp.home_team_id
                and mp.away_team_id = mgp.away_team_id
                and mp.model_id = %s
            join world_cup_matches wm on wm.id = mgp.match_id
            where mgp.goal_model_id = %s
                and mgp.match_date between %s and %s
                and lower(wm.status) in ('scheduled', 'schedule', 'timed')
            order by wm.kickoff_at asc, mgp.match_date asc, mgp.id asc
            """,
            (result_model_id, goal_model_id, from_date, to_date),
        ).fetchall()
    return [dict(row) for row in rows]


def update_calibrated_goal_prediction(
    db: Database,
    *,
    match_goal_prediction_id: str,
    result_model_id: str,
    result_probabilities: dict[str, float],
    summary: dict,
    top_scores: list[dict],
    calibration_method: str,
) -> None:
    with db.connect() as conn:
        conn.execute(
            """
            update match_goal_predictions
            set
                result_model_id = %s,
                expected_home_goals = %s,
                expected_away_goals = %s,
                most_likely_home_score = %s,
                most_likely_away_score = %s,
                over_1_5_probability = %s,
                over_2_5_probability = %s,
                both_teams_score_probability = %s,
                result_probabilities = %s::jsonb,
                calibration_method = %s,
                score_probability_mass = %s,
                calibrated_at = now(),
                source = 'goals-ml-service+result-calibration',
                updated_at = now()
            where id = %s
            """,
            (
                result_model_id,
                summary["expected_home_goals"],
                summary["expected_away_goals"],
                summary["most_likely_home_score"],
                summary["most_likely_away_score"],
                summary["over_1_5_probability"],
                summary["over_2_5_probability"],
                summary["both_teams_score_probability"],
                json.dumps(_json_safe(result_probabilities)),
                calibration_method,
                summary["score_probability_mass"],
                match_goal_prediction_id,
            ),
        )
        conn.execute("delete from match_score_probabilities where match_goal_prediction_id = %s", (match_goal_prediction_id,))
        if top_scores:
            with conn.cursor() as cur:
                cur.executemany(
                    """
                    insert into match_score_probabilities (
                        match_goal_prediction_id, home_score, away_score, probability
                    )
                    values (%s, %s, %s, %s)
                    on conflict (match_goal_prediction_id, home_score, away_score) do update set
                        probability = excluded.probability
                    """,
                    [
                        (
                            match_goal_prediction_id,
                            row["home_score"],
                            row["away_score"],
                            row["probability"],
                        )
                        for row in top_scores
                    ],
                )
        conn.commit()
