CREATE TABLE posts (
    "id"        serial primary key,
    "recip"     text              not null,
    "counter"   integer not null,
    "source"    text              not null,
    "title"     text              not null,
    "body"      text              not null,
    "url"       text              not null,
    "timestamp" timestamp default current_timestamp
);

