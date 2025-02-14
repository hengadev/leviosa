-- +goose Up
-- +goose StatementBegin
INSERT INTO events (
    id, 
    location, 
    placecount, 
    freeplace, 
    beginat, 
    sessionduration, 
    priceid, 
    day, 
    month, 
    year
    ) VALUES
        (
            'b16a6f38-d2fb-428c-b97c-929b1010b951', 
            'Impasse Inconnue', 
            23, 
            19, 
            '08:00:00', 
            30, 
            'v0ersgF_de_+wf', 
            13, 
            7, 
            1998
        ),
        (
            '9a676c5d-c9ec-4266-a426-24e5d4983caf', 
            'Impasse Inconnue', 
            22, 
            0, 
            '08:00:00', 
            30, 
            'f9eg93edw0HFur', 
            14, 
            7, 
            1998
        );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM events WHERE rowid > 2;
-- +goose StatementEnd
