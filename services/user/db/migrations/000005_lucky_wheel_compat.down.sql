DELETE FROM "achievement" WHERE code IN ('first_spin', 'lucky_wheel_winner', 'streak_six');

ALTER TABLE "client_profile" DROP COLUMN IF EXISTS last_wheel_spin_at;
ALTER TABLE "client_profile" DROP COLUMN IF EXISTS streak_freeze_active;
