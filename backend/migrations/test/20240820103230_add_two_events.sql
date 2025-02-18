-- +goose Up
-- +goose StatementBegin
INSERT INTO events (
    id,
    encrypted_title,
    encrypted_description,
    encrypted_city,
    encrypted_postal_code, encrypted_address1,
    encrypted_address2,
    placecount,
    freeplace,
    encrypted_begin_at,
    encrypted_end_at,
    encrypted_price_id,
    day,
    month,
    year
    ) VALUES
        (
            '43391431-984f-4b06-8fcc-88040215430b',
            'Second event for Leviosa',
            'Second description for Leviosa',
            'Marseille',
            'postalCode2',
            'address1 - 2',
            '',
            6,
            3,
            '09:00:00',
            '19:00:00',
            'bdab8511-875a-46d5-a228-6db7aecb42a2',
            17,
            5,
            2025
        ),
        (
            '9a676c5d-c9ec-4266-a426-24e5d4983caf',
            'Third event for Leviosa',
            'Third description for Leviosa',
            'Lyon',
            'postalCode3',
            'address1 - 3',
            '',
            18,
            4,
            '10:00:00',
            '32:00:00',
            'ef55b80d-6eb2-4e22-9b68-ea219c202c71', 
            3,
            6,
            2025
        );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM events WHERE rowid > 2;
-- +goose StatementEnd
