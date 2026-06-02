create table servers (
    server_id text primary key,
    server_name text not null unique,
    ipv4 text not null,
    status text not null,
    created_at timestamp,
    metadata_updated_at timestamp,
    last_ping_at timestamp,
    is_deleted boolean
);

create table users (
    id serial primary key,
    full_name varchar(255) not null,
    email varchar(255) not null unique
);

create table accounts (
    id serial primary key,
    username varchar(255) not null unique,
    password varchar(255) not null,
    role varchar(50) not null,
    user_id int not null,
    foreign key (user_id) references users(id) on delete cascade
);

insert into users (full_name, email)
values ('Hari', 'le31052003@gmail.com');

insert into accounts (username, password, role, user_id)
values (
    'hari',
    '$2a$12$0.tjaJ0/aJPAG311QCIG3OKZIeUnElkv1savvhX0PtqLlqb/Cb9dq',
    'admin',
    1
);