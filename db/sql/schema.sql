CREATE TABLE student
(
    roll_number     text primary key,
    name            text,
    fathers_name    text,
    batch           text,
    branch          text,
    latest_semester integer,
    cgpi            real
);

CREATE TABLE subject_result_data
(
    student_roll_number text,
    semester            integer,
    subject_code        text,
    grade               text,
    sub_gp              int,
    primary key (student_roll_number, subject_code),
    foreign key (subject_code) references subject (code),
    foreign key (student_roll_number) references student (roll_number),
    foreign key (subject_code) references subject (code)
);

CREATE TABLE semester_result_data
(
    student_roll_number text,
    semester            integer,
    cgpi                text,
    sgpi                text,
    primary key (student_roll_number, semester),
    foreign key (student_roll_number) references student (roll_number)
);

CREATE TABLE subject
(
    code    text primary key,
    name    text,
    credits integer
);