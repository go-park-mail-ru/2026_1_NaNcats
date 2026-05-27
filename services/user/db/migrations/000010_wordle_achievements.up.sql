-- Ачивки за «5 букв»: первая победа, 10 побед всего, 30 побед подряд.

INSERT INTO "achievement" (code, title, description, icon, sort_order) VALUES
    ('wordle_first_win', 'Слово найдено', 'Отгадайте слово дня в «5 букв» в первый раз', '📝', 7),
    ('wordle_winner_10', 'Лексикон', 'Отгадайте 10 слов дня в «5 букв»', '📚', 8),
    ('wordle_streak_30', 'Марафонец', 'Отгадывайте слово дня 30 дней подряд', '🏆', 9)
ON CONFLICT (code) DO NOTHING;
