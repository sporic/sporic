create table user(
    user_id integer primary key auto_increment, -- internal user id
    username varchar(255) not null, -- faculty id that 4, 5, 6 digit number
    hashed_password char(60) not null,
    email varchar(255) not null,
    created_at datetime not null,
    user_role integer not null
);
alter table user
add constraint user_uc_username unique (username);