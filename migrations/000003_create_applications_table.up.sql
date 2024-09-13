create table applications(
    sporic_ref_no varchar(255) primary key,
    leader integer not null,
    financial_year integer not null,
    activity_type integer not null,
    estimated_amt integer not null,
    company_name varchar(255) not null,
    company_adress varchar(255) not null,
    contact_person_name varchar(255) not null,
    contact_person_designation varchar(255) not null,
    contact_person_email varchar(255) not null,
    contact_person_mobile varchar(10) not null,
    project_status integer not null,
    project_start_date datetime,
    project_end_date datetime,
    foreign key (leader) references user(user_id)
);