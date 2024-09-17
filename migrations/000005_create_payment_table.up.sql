create table payment(
    payment_id integer primary key auto_increment,
    sporic_ref_no varchar(255) not null,
    payment_amt integer not null,
    gst_number varchar(255),
    pan_number varchar(255),
    payment_date datetime,
    payment_status integer not null,
    foreign key (sporic_ref_no) references applications(sporic_ref_no)
);