create table expenditure(
    expenditure_id integer primary key auto_increment,
    sporic_ref_no varchar(255) not null,
    expenditure_name varchar(255) not null,
    expenditure_amt integer not null,
    expenditure_date datetime not null,
    expenditure_status integer not null,
    foreign key (sporic_ref_no) references applications(sporic_ref_no)
);