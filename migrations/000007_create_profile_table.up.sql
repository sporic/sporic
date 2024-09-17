create table profile (
    profile_id integer primary key auto_increment,
    user_id integer not null,
    username varchar(255) not null,
    full_name varchar(255) not null,
    designation varchar(255) not null,
    mobile_number varchar(255) not null,
    email varchar(255) not null,
    school varchar(255) not null,
    foreign key (user_id) references user(user_id)
);