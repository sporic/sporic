create table user(
    user_id integer primary key auto_increment,
    username varchar(255) not null,
    hashed_password char(60) not null,
    email varchar(255) not null,
    created_at datetime not null,
    user_role integer not null
);
alter table user
add constraint user_uc_username unique (username);