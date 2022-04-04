package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	memorycache "github.com/maxchagin/go-memorycache-example"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	validator "gopkg.in/go-playground/validator.v9"
)

type Service struct {
	Ch   *memorycache.Cache
	Stan struct {
		Conn stan.Conn
		Sub  stan.Subscription
	}
	Db *sql.DB
	QR QR
}

type QR struct {
	order      *sql.Row
	orderitems *sql.Rows
	items      *sql.Rows
	del_id     int
	pay_id     string
	items_id   pq.Int32Array
	delivery   *sql.Row
	payment    *sql.Row
}

func (service *Service) orderHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpl, err := template.ParseFiles("interface.html")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	case "POST":
		//var err error
		order_uid := r.PostFormValue("order_uid")
		v, found := (service.Ch.Get(order_uid))
		//log.Println(found)
		if found == false {
			fmt.Fprintf(w, "DB:\n")
			var ret Model
			ret.OrderUid = order_uid
			var err error
			err = service.SelectOrder(&ret)
			if err != nil {
				return
			}

			err = service.SelectItems(&ret)
			if err != nil {
				return
			}

			err = service.SelectDelivery(&ret)
			if err != nil {
				return
			}

			err = service.SelectPayment(&ret)
			if err != nil {
				return
			}

			service.Ch.Set(order_uid, ret, 0)
			v, _ = (service.Ch.Get(order_uid))
		} else {
			fmt.Fprintf(w, "CACHE:\n")
		}
		b, err := json.MarshalIndent(v, "", "\t")
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Fprintf(w, string(b))
		return
	}
}

func (service *Service) msgHandler(m *stan.Msg) {

	var obj Model
	err := json.Unmarshal(m.Data, &obj)
	if err != nil {
		log.Println(err)
		return
	}

	v := validator.New()
	err = v.Struct(obj)
	if err != nil {
		log.Println(err)
		return
	}
	service.Ch.Set(obj.OrderUid, obj, 0)

	tx, err := service.Db.Begin()
	var delid int
	items := make([]int, len(obj.Items))
	for i, e := range obj.Items {
		items[i] = e.ChrtId
	}

	err = tx.QueryRow(`	INSERT INTO public.delivery ("name",phone,zip,city,address,region,email)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`,
		obj.Delivery.Name,
		obj.Delivery.Phone,
		obj.Delivery.Zip,
		obj.Delivery.City,
		obj.Delivery.Address,
		obj.Delivery.Region,
		obj.Delivery.Email,
	).Scan(&delid)
	_, err = tx.Exec(`INSERT INTO public.orders (order_uid,track_number,entry,delivery,payment,items,locale,internal_signature,customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`,
		obj.OrderUid,
		obj.TrackNumber,
		obj.Entry,
		delid,
		obj.Payment.Transaction,
		pq.Array(items),
		obj.Locale,
		obj.InternalSignature,
		obj.CustomerId,
		obj.DeliveryService,
		obj.Shardkey,
		obj.SmId,
		obj.DateCreated,
		obj.OofShard)

	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}
	tx.Commit()
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	} else {
		log.Printf("Added Order UID: %s\n", obj.OrderUid)
	}

}

func main() {

	DBconnStr := "host=" + os.Getenv("POSTGRES_NAME") + " user=" + os.Getenv("POSTGRES_USER") + " dbname=" + os.Getenv("POSTGRES_USER") + " password=" + os.Getenv("POSTGRES_PASSWORD") + " sslmode=disable"

	var err error
	service := &Service{}

	//connect to SERVER
	nc, err := nats.Connect(os.Getenv("STAN_NAME"))
	//init STAN
	service.Stan.Conn, err = stan.Connect(os.Getenv("CLUSTER_ID"), os.Getenv("CLIENT_ID"), stan.NatsConn(nc))
	if err != nil {
		log.Println(err.Error())
		return
	}
	//Subscriber
	service.Stan.Sub, err = service.Stan.Conn.Subscribe(os.Getenv("STAN_GROUP"), service.msgHandler)

	// Close connection
	defer service.Stan.Conn.Close()
	defer service.Stan.Sub.Unsubscribe()
	if err != nil {
		log.Println(err.Error())
		return
	}

	//Database
	log.Println(DBconnStr)
	service.Db, err = sql.Open("postgres", DBconnStr)
	if err != nil {
		log.Fatal(err)
	}

	// init cache
	service.InitCache()

	http.HandleFunc("/order", service.orderHandler)
	http.ListenAndServe(":8090", nil)
}

func (service *Service) InitCache() {
	expiration := 1 * time.Minute
	duration := 1 * time.Minute
	service.Ch = memorycache.New(expiration, duration)

	row, err := service.Db.Query(`SELECT order_uid, track_number, entry, delivery, payment, items, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	FROM public.orders
	LIMIT $1;`, 5)
	if err != nil {
		log.Println(err)
	}
	for row.Next() {
		var ret Model
		row.Scan(
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

		err = service.SelectItems(&ret)

		if err != nil {
			return
		}

		err = service.SelectDelivery(&ret)
		if err != nil {
			return
		}

		err = service.SelectPayment(&ret)
		if err != nil {
			return
		}

		service.Ch.Set(ret.OrderUid, ret, 0)
	}
}
