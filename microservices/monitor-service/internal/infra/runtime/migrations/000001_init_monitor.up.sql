create table monitored_servers (
    server_id text primary key,
    server_name text not null unique,
    ipv4 text not null,
    version bigint default 1
);

create table live_statuses (
    server_id text primary key,
    status varchar(20) not null default 'UNKNOWN',
    last_ping_at timestamp,
    last_heartbeat_at timestamp
);
