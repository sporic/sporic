create table notifications (
    craeted_at datetime not null,
    notification_type integer not null,
    notification_description varchar(255) not null,
    notification_to varchar(255) not null
)