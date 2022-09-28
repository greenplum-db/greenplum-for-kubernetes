-- testing regression that will break if gp_segment_configuration has "127.0.0.1" instead of "master"

select version();
explain select 1;

drop table if exists shipped cascade;

create table shipped (
    value       float8
);

insert into shipped values (1.0);

insert into shipped (value) values ((select value from shipped));

select * from shipped;
