package main
import (
    "fmt"
    "time"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

func PrepareTx(db *sql.DB,qry string) (tx *sql.Tx, s *sql.Stmt, e error) {
 if tx,e=db.Begin(); e!=nil {
  return
 }

 if s, e = tx.Prepare(qry);e!=nil {
	 panic(e.Error())
 }
 return
}

func main() {
    done := false
    var i int64
    query := "insert into customer (email) values (?)"
    i = 0
    db, err := sql.Open("mysql", "root@tcp(127.0.0.1:15306)/test_sharded_keyspace")
    fmt.Println("connection opened")
    if err != nil {
	panic(err.Error())
     }
    defer db.Close()
    now := time.Now().UnixNano()
    tx, stmt, err := PrepareTx(db, query)
    if err != nil {
	panic(err.Error())
    }
    for !done {
		i++
	    time.Sleep(1 * time.Millisecond)
		if _,err := stmt.Exec(fmt.Sprintf("a%d",now+i)); err != nil {
				panic(err)
		}
		if i % 1000  == 0 {
			if err := tx.Commit(); err!=nil {
				panic(err)
			}
			tx, stmt, err = PrepareTx(db, query)
			if err != nil {
				panic(err.Error())
			}
		}
    }
}