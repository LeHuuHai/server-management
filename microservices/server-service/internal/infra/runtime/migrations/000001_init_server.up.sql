create table server_profiles (
    server_id text primary key,
    server_name text not null unique,
    ipv4 text not null,
    created_at timestamp,
    updated_at timestamp,
    version bigint default 1,
    is_deleted boolean default false
);

create table outbox_events (
    id varchar(36) primary key,
    topic varchar(255) not null,
    payload jsonb not null,
    status varchar(50) not null default 'PENDING',
    created_at timestamp not null
);
