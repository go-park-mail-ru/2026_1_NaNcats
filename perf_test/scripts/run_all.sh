#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

bash 00_prereqs.sh
bash 01_infra_up.sh
bash 02_migrate.sh
bash 03_services_up.sh
bash 04_provision_owner.sh

# --- Итерация 0: baseline (без перф-индексов) ---
bash 05_load_create.sh baseline      # залить 100k через эндпоинт создания
bash explain.sh        baseline
bash 06_load_read.sh   baseline       # чтение search + list

# --- Итерация 1: триграммные индексы (оптимизация поиска) ---
bash 07_optimize.sh    9
bash explain.sh        opt9
WORKERS=20 bash 06_load_read.sh opt    # перезамер (search улучшится)

# --- Итерация 2: btree(promotion_tier DESC, id) (оптимизация листинга) ---
bash 07_optimize.sh    10
bash explain.sh        opt
WORKERS=20 bash 06_load_read.sh opt    # перезамер list (перезапишет opt)

# --- Доп.: стоимость записи при наличии индексов ---
N=5000 bash 05_load_create.sh withindex || true

echo ">> готово. Смотри perf_test/results/ и perf_test/README.md"
