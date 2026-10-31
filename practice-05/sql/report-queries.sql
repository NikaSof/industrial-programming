-- Query Tool базы todo. Выполнять выделенный блок, а не весь файл.

-- 01-database.png: окружение и текущая база
SELECT current_database(), version();

-- 01-schema.png: столбцы таблицы
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'tasks'
ORDER BY ordinal_position;

-- 04-select.png: записи после создания через API
SELECT id, title, done, created_at, updated_at FROM public.tasks ORDER BY id;

-- 09-seed-select.png: результат go run . -seed
SELECT id, title, done FROM public.tasks
WHERE title IN ('Подготовить практику 5', 'Проверить транзакцию', 'Оформить README')
ORDER BY id;

-- Необязательный опыт отката. Выполнить этот блок отдельно.
BEGIN;
INSERT INTO public.tasks(title) VALUES ('rollback-screenshot-pr5');
-- Отдельно выполнить следующую команду: ожидается ошибка NOT NULL.
INSERT INTO public.tasks(title) VALUES (NULL);
-- После ошибки обязательно выполнить отдельно:
ROLLBACK;
-- 10-rollback.png: ожидается отсутствие строки, если её не было до опыта.
SELECT * FROM public.tasks WHERE title = 'rollback-screenshot-pr5';
