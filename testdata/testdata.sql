insert into users (username) values 
('alice'),
('bob');

insert into frozen_balances (userid, amt, status, idempotencykey) values
(1, 10, 1, 'transaction1'),
(1, 20, 2, 'transaction2'),
(2, 15, 1, 'transaction3'),
(4, 25, 2, 'transaction4');

insert into transactions (id, userid, amt, isdebit) values
('transaction1', 1, 100, false),
('transaction2', 2, 200, false),
('transaction3', 1, 130, false),
('transaction4', 2, 140, false);
