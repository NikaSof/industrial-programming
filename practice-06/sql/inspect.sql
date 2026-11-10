-- Query Tool базы pz6_gorm; выполнять нужный SELECT отдельно.
SELECT current_database(), version();
SELECT * FROM users ORDER BY id;
SELECT * FROM notes ORDER BY id;
SELECT * FROM tags ORDER BY id;
SELECT * FROM note_tags ORDER BY note_id, tag_id;

SELECT n.id, n.title, u.name AS author, t.name AS tag
FROM notes n
JOIN users u ON u.id = n.user_id
LEFT JOIN note_tags nt ON nt.note_id = n.id
LEFT JOIN tags t ON t.id = nt.tag_id
ORDER BY n.id, t.name;
