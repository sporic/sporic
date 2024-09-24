alter table payment
add column currency varchar(3) not null,
    add column tax integer not null;