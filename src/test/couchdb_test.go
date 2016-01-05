// CouchDB_test

package test

import (
	"encoding/json"
	"flag"
	"fmt"
	// 	caes_json "github.com/carneades/carneades-4/src/engine/caes/encoding/json"
	"github.com/fjl/go-couchdb"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

type Welcome struct {
	Couchdb string       `json:"couchdb"`
	Uuid    string       `json:"uuid"`
	Version string       `json:"version"`
	Vendor  VendorStruct `json:"vendor"`
}

type VendorStruct struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type UuidsStruct struct {
	Uuids []string `json:"uuids"`
}

//var even_loop = caes_json.TempArgGraphDB{Meta: map[string]interface{}{"title": "Even Loop Example"},
//	Statements: map[string]caes_json.TempStatement{"P": {Text: "P"}, "Q": {Text: "Q"}},
//	Arguments: map[string]caes_json.TempArgument{"a1": {Premises: []interface{}{"P"}, Conclusion: "Q"},
//		"a2": {Premises: []interface{}{"Q"}, Conclusion: "P"}}}

func TestCouchDB(t *testing.T) {
	flag.Parse()
	resp, err := http.Get("http://127.0.0.1:5984/")
	if err != nil {
		log.Fatal(" ***Error: http.Get(\"http://127.0.0.1:5984/:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(" ***Error: ioutil.ReadAll(resp.Body):", err)
	}
	fmt.Printf("> http://127.0.0.1:5984/ \n< %s \n\n", body)
	var welcome Welcome
	err = json.Unmarshal(body, &welcome)
	if err != nil {
		log.Fatal(" ***Error: json.Unmarshal(body, &welcome):", err)
	}
	fmt.Printf("Unmarchal: %v\n", welcome)

	fmt.Printf("\n Liste aller Datenbanken: \n")
	//             -----------------------
	resp, err = http.Get("http://127.0.0.1:5984/_all_dbs")
	if err != nil {
		log.Fatal(" ***Error: http.Get(\"http://127.0.0.1:5984/:", err)
	}
	defer resp.Body.Close()
	// fmt.Printf(" Type: %T \nRep: \n%v\n\n", resp, resp)
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(" ***Error: ioutil.ReadAll(resp.Body):", err)
	}
	fmt.Printf("> http://127.0.0.1:5984/_all_dbs \n< %s \n\n", body)

	// Test github.com/fjl/go-couchdb
	// ------------------------------

	c, err := couchdb.NewClient("http://127.0.0.1:5984", nil)
	if err != nil {
		log.Fatal(" ***Error: couchdb.NewClient(\"http://127.0.0.1:5984\", nil):", err)
	}
	fmt.Printf("\n Type of c: %T \n Value of c: %v\n", c, c)
	err = c.Ping()
	if err != nil {
		log.Fatal(" ***Error: c.Ping():", err)
	}
	fmt.Printf("> Ping \"http://127.0.0.1:5984\" < OK ")

	fmt.Printf("\n Neue Datenbank anlegen: \n")
	//             ---------------------------

	// db, err := c.CreateDB("testdb")
	db, err := c.EnsureDB("testdb")
	// DB-namen dürfen nur die folgenden Zeichen enthalten: (a-z),(0-9),_,(,),+,-,/
	if err != nil {
		log.Fatal(" *Error: c.CreateDB(\"testDb\"):", err)
	}

	fmt.Printf(" Type db: %T\n db: %v\n", db, db)

	fmt.Printf("\n Get new UUID's \n")
	//          ---------------------------

	// get a new UUID's
	// ----------------
	// resp, err = http.Get("http://127.0.0.1:5984/_uuids")
	// get 10 new UUID
	resp, err = http.Get("http://127.0.0.1:5984/_uuids?count=10")

	if err != nil {
		log.Fatal(" ***Error: http.Get(\"http://127.0.0.1:5984/_uuids:", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(" ***Error: [_uuid] ioutil.ReadAll(resp.Body):", err)
	}
	fmt.Printf("> http://127.0.0.1:5984/_uuids \n< %s \n\n", body)
	var uuids UuidsStruct
	err = json.Unmarshal(body, &uuids)
	if err != nil {
		log.Fatal(" ***Error: json.Unmarshal(body, &uuids):", err)
	}
	fmt.Printf("Unmarchal: %v\n[]uuids: %v", uuids, uuids.Uuids)
	// 	uuidIdx := 0

	//	fmt.Printf("\n Create a new Dokument \n")
	//	//          ---------------------------
	//	_, err = db.Put(uuids.Uuids[uuidIdx], even_loop, "")
	//	check(t, err)

}
