-- +goose Up
-- +goose StatementBegin
INSERT INTO message_templates(template_id, code, channel, subject, body, is_active)
VALUES
    (
        gen_random_uuid(),
        'intern_invite_v1',
        'email',
        'X5 Group: приглашение на следующий этап',
        $$Здравствуйте, {{.first_name}}!

        Спасибо за отклик на стажировку X5 Group.
        Мы приглашаем вас на следующий этап. В ближайшее время HR свяжется с вами.

        С уважением,
        X5 Group$$,
        true
    ),
    (
        gen_random_uuid(),
        'intern_reject_v1',
        'email',
        'X5 Group: результат по стажировке',
        $$Здравствуйте, {{.first_name}}!

        Спасибо за интерес к стажировке X5 Group.
        К сожалению, на текущем этапе мы не готовы продолжить процесс.

С уважением,
        X5 Group$$,
        true
    )
    ON CONFLICT (code) DO UPDATE
                              SET channel   = EXCLUDED.channel,
                              subject   = EXCLUDED.subject,
                              body      = EXCLUDED.body,
                              is_active = EXCLUDED.is_active,
                              updated_at = now();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM message_templates
WHERE code IN ('intern_invite_v1', 'intern_reject_v1');
-- +goose StatementEnd
