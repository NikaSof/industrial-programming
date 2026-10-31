--окружение и текущая база
SELECT current_database(), version();

-- столбцы таблицы
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'tasks'
ORDER BY ordinal_position;

-- записи после создания через API
SELECT id, title, done, created_at, updated_at FROM public.tasks ORDER BY id;

-- результат go run . -seed
SELECT id, title, done FROM public.tasks
WHERE title IN ('Подготовить практику 5', 'Проверить транзакцию', 'Оформить README')
ORDER BY id;

BEGIN;
INSERT INTO public.tasks(title) VALUES ('rollback-screenshot-pr5');
INSERT INTO public.tasks(title) VALUES (NULL);

ROLLBACK;

SELECT * FROM public.tasks WHERE title = 'rollback-screenshot-pr5';
