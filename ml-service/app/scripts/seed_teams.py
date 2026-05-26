from __future__ import annotations

import argparse
import logging
import sys
from pathlib import Path

from dotenv import load_dotenv

sys.path.append(str(Path(__file__).resolve().parents[2]))

from app.metrics.database import Database
from app.metrics.local_data import load_alias_config


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--aliases-file", default=None)
    return parser.parse_args()


def main() -> None:
    load_dotenv()
    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(name)s: %(message)s")
    args = parse_args()

    db = Database()
    summary = db.seed_teams_and_aliases(load_alias_config(args.aliases_file))

    print(
        "teams seed completed: "
        f"teams_inserted={summary['teams_inserted']} teams_updated={summary['teams_updated']} "
        f"aliases_inserted={summary['aliases_inserted']} aliases_updated={summary['aliases_updated']}"
    )


if __name__ == "__main__":
    main()
