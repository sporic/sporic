create table teams(
    sporic_ref_no varchar(255),
    team_member varchar(255),
    foreign key (sporic_ref_no) references applications(sporic_ref_no)
);