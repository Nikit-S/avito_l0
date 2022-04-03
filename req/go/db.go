package main

import (
	"log"

	"github.com/lib/pq"
)

func (service *Service) SelectPayment(ret *Model) error {
	service.QR.payment = service.Db.QueryRow(
		`SELECT 
		"transaction", 
		request_id, 
		currency, 
		provider, 
		amount, 
		payment_dt, 
		bank, 
		delivery_cost, 
		goods_total, 
		custom_fee
		FROM public.payment
		WHERE "transaction"=$1;`, service.QR.pay_id)

	err := service.QR.payment.Scan(&ret.Payment.Transaction,
		&ret.Payment.RequestId,
		&ret.Payment.Currency,
		&ret.Payment.Provider,
		&ret.Payment.Amount,
		&ret.Payment.PaymentDt,
		&ret.Payment.Bank,
		&ret.Payment.DeliveryCost,
		&ret.Payment.GoodsTotal,
		&ret.Payment.CustomFee)
	return err
}

func (service *Service) SelectDelivery(ret *Model) error {
	service.QR.delivery = service.Db.QueryRow(
		`SELECT 
		"name", 
		phone, 
		zip, 
		city, 
		address, 
		region, 
		email
		FROM public.delivery
		WHERE id=$1;`, service.QR.del_id)

	err := service.QR.delivery.Scan(&ret.Delivery.Name,
		&ret.Delivery.Phone,
		&ret.Delivery.Zip,
		&ret.Delivery.City,
		&ret.Delivery.Address,
		&ret.Delivery.Region,
		&ret.Delivery.Email)
	return err
}

func (service *Service) SelectItems(ret *Model) error {
	stmt, err := service.Db.Prepare(`SELECT 
			chrt_id, 
			track_number, 
			price, 
			rid, 
			name, 
			sale, 
			size, 
			total_price, 
			nm_id, 
			brand, 
			status
			FROM public.items
			WHERE chrt_id = ANY ($1);`)

	service.QR.items, err = stmt.Query(pq.Array(service.QR.items_id))
	if err != nil {
		return err
	}
	arr2 := make([]Item, 0)
	var i Item
	for service.QR.items.Next() {
		service.QR.items.Scan(&i.ChrtId,
			&i.TrackNumber,
			&i.Price,
			&i.Rid,
			&i.Name,
			&i.Sale,
			&i.Size,
			&i.Total_price,
			&i.Nm_id,
			&i.Brand,
			&i.Status)
		arr2 = append(arr2, i)
	}
	ret.Items = arr2
	service.QR.items.Close()
	return err
}

/*func (service *Service) SelectOrderItems(ret *Model) (*[]int, error) {
	var err error
	service.QR.orderitems, err = service.Db.Query(
		`SELECT
		item
		FROM public.orderitems
		WHERE order_uid = $1;`, ret.OrderUid)

	if err != nil {
		return nil, err
	}
	arr := make([]int, 0)
	var t int
	for service.QR.orderitems.Next() {
		service.QR.orderitems.Scan(&t)
		log.Println("t:", t)
		arr = append(arr, t)
	}
	service.QR.orderitems.Close()
	return &arr, err
}*/

func (service *Service) SelectOrder(ret *Model) error {
	service.QR.order = service.Db.QueryRow(
		`SELECT 
		order_uid, 
		track_number, 
		entry, 
		delivery, 
		payment, 
		items,
		locale, 
		internal_signature, 
		customer_id, 
		delivery_service, 
		shardkey, 
		sm_id, 
		date_created, 
		oof_shard 
		FROM orders 
		WHERE order_uid = $1`, ret.OrderUid)

	err := service.QR.order.Scan(
		&ret.OrderUid,
		&ret.TrackNumber,
		&ret.Entry,
		&service.QR.del_id,
		&service.QR.pay_id,
		&service.QR.items_id,
		&ret.Locale,
		&ret.InternalSignature,
		&ret.CustomerId,
		&ret.DeliveryService,
		&ret.Shardkey,
		&ret.SmId,
		&ret.DateCreated,
		&ret.OofShard)
	log.Println(service.QR.items_id)
	return err
}
