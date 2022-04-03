
CREATE TABLE public.delivery (
	id SERIAL,
	name varchar,
	phone varchar,
	zip varchar,
	city varchar,
	address varchar,
	region varchar,
	email varchar,
	primary key (id)
);

CREATE TABLE public.payment (
	transaction varchar,
	request_id varchar,
	currency varchar,
	provider varchar,
	amount integer,
	payment_dt integer,
	bank varchar,
	delivery_cost integer,
	goods_total integer,
	custom_fee integer,
	primary KEY(transaction)
);

CREATE TABLE public.items (
	chrt_id integer,
	track_number varchar,
	price integer,
	rid varchar,
	name varchar,
	sale integer,
	size varchar,
	total_price integer,
	nm_id integer,
	brand varchar,
	status integer,
	primary KEY(chrt_id)
);

CREATE TABLE public.orders (
	order_uid varchar,
	track_number varchar,
	entry varchar,
	delivery integer references delivery(id),
	payment varchar references payment(transaction),
	items integer[],
	locale varchar,
	internal_signature varchar,
	customer_id varchar,
	delivery_service varchar,
	shardkey varchar,
	sm_id integer, 
	date_created timestamp,
	oof_shard varchar,
	primary KEY(order_uid)
);

--CREATE TABLE public.orderitems (
--	order_uid varchar references orders(order_uid),
--	item integer references items(chrt_id)
--);


-- Auto-generated SQL script #202204012202
INSERT INTO public.items (chrt_id,track_number,price,rid,"name",sale,"size",total_price,nm_id,brand,status)
	VALUES (9934930,'WBILMTESTTRACK',453,'ab4219087a764ae0btest','Mascaras',30,'0',317,2389212,'Vivienne Sabo',202),
	(1,'A',500,'ab4219087a764ae0btest','E',30,'0',317,2389212,'D',202),
	(2,'B',600,'ab4219087a764ae0btest','C',30,'0',317,2389212,'F',202);

-- Auto-generated SQL script #202204012209
INSERT INTO public.payment ("transaction",request_id,currency,provider,amount,payment_dt,bank,delivery_cost,goods_total,custom_fee)
	VALUES ('b563feb7b2b84b6test','','USD','wbpay',1817,1637907727,'alpha',1500,317,0),
	('b563feb1b84b6test','','RUB','sber',1,1637907726,'sber',100,100,0),
	('b563feb734b6test','','DOH','wtf',2,1637907725,'tink',200,200,0);

-- Auto-generated SQL script #202204012212
INSERT INTO public.delivery ("name",phone,zip,city,address,region,email)
	VALUES ('Test Testov','+9720000000','2639809','Kiryat Mozkin','Ploshad Mira 15','Kraiot','test@gmail.com'),
	('barcher','+972012300','12','barcher','moscow','russia','noemail');

-- Auto-generated SQL script #202204012214
INSERT INTO public.orders (order_uid,track_number,entry,delivery,payment,locale,internal_signature,customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard,items)
	VALUES ('b563feb7b2b84b6test','WBILMTESTTRACK','WBIL',1,'b563feb7b2b84b6test','en','','test','meest','9',99,'2021-11-26T06:22:19Z','1','{9934930}');

-- Auto-generated SQL script #202204012215
--INSERT INTO public.orderitems (order_uid,item)
--	VALUES ('b563feb7b2b84b6test',9934930);

