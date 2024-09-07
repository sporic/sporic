create table applications(
    sporic_ref_no varchar(255) primary key,
    financial_year varchar(9) not null,
    activity_type varchar(255) not null,
    leader varchar(255) not null,
    estimated_amt integer not null,
    company_name varchar(255) not null,
    company_adress varchar(255) not null,
    contact_person varchar(255) not null,
    mail_id varchar(255) not null,
    mobile varchar(10) not null,
    gst varchar(255) not null,
    pan_number varchar(255) not null,
    status integer not null
);