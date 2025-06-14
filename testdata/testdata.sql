SET timezone to 'Asia/Singapore';

delete from users;
insert into users (id, username) values 
(1, 'alice'),
(2, 'bob');

delete from frozen_balances;
insert into frozen_balances (idempotencykey, userid, otherid, amt, status, at) values
('1to2frozen', 1, 2, 10, 'frozen',  now()),
('1to2confirmed', 1, 2, 20, 'confirmed',  now()), -- 1 transfer to 2 20 units confirmed
('2to1fronzen', 2, 1, 15, 'frozen',  now()),
('2to1confirmed', 2, 1, 25, 'confirmed',  now()); -- 2 transfer to 1 25 units confirmed

delete from transactions;
insert into transactions (idempotencykey, userid, otherid, amt, isdebit, at) values
('1to2confirmed', 1, 2, 20, true, now()),
('1to2confirmed', 2, 1, 20, false, now()),
('2to1confirmed', 2, 1, 25, true, now()),
('2to1confirmed', 1, 2, 25, false, now()),
('user1credit', 1, null, 100, false, now()),
('user1debit', 1, null, 5, true, now()),
('user2credit', 2, null, 130, false, now()),
('user2debit', 2, null, 10, true, now());
