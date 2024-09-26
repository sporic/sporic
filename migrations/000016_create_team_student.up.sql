create table team_student(
    sporic_ref_no varchar(255),
    member_name varchar(255),
    foreign key (sporic_ref_no) references applications(sporic_ref_no)
);                      