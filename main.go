package main

import (
	"fmt"
	"log"
	"time"

	"github.com/atbys/gabby"
	"github.com/jinzhu/gorm"
	gs "github.com/soniah/gosnmp"
)

type Info struct {
	Ipaddr  string
	Macaddr string
	Port    int
}

func Connect() (*gorm.DB, error) {
	DBMS := "postgres"
	USER := "user=gabby"
	//PASS := ""
	HOST := "host=127.0.0.1"
	PORT := "port=5432"
	DBNAME := "dbname=network_test"
	SSLMODE := "sslmode=disable"
	CONNECT := HOST + " " + PORT + " " + USER + " " + DBNAME + " " + SSLMODE

	db, err := gorm.Open(DBMS, CONNECT)
	if err != nin {
		return nil, err
	}

	return db, nil
}

func InsertInfo(ip string, mac string, port int) error {
	db, err := Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	var info Info
	info.Ipaddr = ip
	info.Macaddr = mac
	info.Port = port
	db.Table("").Create(&info)

	return nil
}

func GetPortNum(mac string) {
	params := &gs.GoSNMP{
		Target:        "192.168.",
		Port:          161,
		Version:       g.Version3,
		Timeout:       time.Duration(30) * time.Second,
		SecurityModel: g.UserSecurityModel,
		MsgFlags:      g.AuthPriv,
		SecurityParameters: &g.UsmSecurityParameters{UserName: "user",
			AuthenticationProtocol:   g.SHA,
			AuthenticationPassphrase: "password",
			PrivacyProtocol:          g.DES,
			PrivacyPassphrase:        "password",
		},
	}

	err := params.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer params.Conn.Close()

	oids := []string{"1.3.6.1.2.1.17.7.1.2", "1.3.6.1.2.1.1.7.0"}
	result, err2 := params.Get(oids) // Get() accepts up to g.MAX_OIDS
	if err2 != nil {
		log.Fatalf("Get() err: %v", err2)
	}

	for i, variable := range result.Variables {
		fmt.Printf("%d: oid: %s ", i, variable.Name)

		// the Value of each variable returned by Get() implements
		// interface{}. You could do a type switch...
		switch variable.Type {
		case g.OctetString:
			fmt.Printf("string: %s\n", string(variable.Value.([]byte)))
		default:
			// ... or often you're just interested in numeric values.
			// ToBigInt() will return the Value as a BigInt, for plugging
			// into your calculations.
			fmt.Printf("number: %d\n", g.ToBigInt(variable.Value))
		}
	}
}

func RequestFromRouter(c *gabby.Context) {
	if c.State == gabby.USED {
		InsertInfo(c.DstIPaddr.String(), c.DstMACaddr.String(), 0)
	}
	return
}

func RequestFromHost(c *gabby.Context) {
	InsertInfo(c.SrcIPaddr.String(), c.SrcMACaddr.String(), 0)
	if c.State == gabby.USED {
		InsertInfo(c.DstIPaddr.String(), c.DstMACaddr.String(), 0)
	}
	return
}

func Used(c *gabby.Context) {
	InsertInfo(c.SrcIPaddr.String(), c.SrcMACaddr.String(), 0)
	return
}

func main() {
	e, err := gabby.Default()
	if err != nil {
		fmt.Println("fault")
		return
	}

	e.RegistHandle(gabby.REQUEST_FROM_ROUTER, RequestFromRouter)
	e.RegistHandle(gabby.REQUEST_FROM_HOST, RequestFromHost)
	e.RegistHandle(gabby.USED_PACKET, Used)

	e.Run()
}
