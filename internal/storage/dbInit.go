package storage



import (
	"database/sql"
	// "log"

	_ "modernc.org/sqlite"
)

// type documentdb struct{
// 	id int
// 	url string	
// 	title string
// 	body string
// 	createdAt string
// }

func OpenDbInstance(path string) (*sql.DB,error){
	db,err:= sql.Open("sqlite",path)
	if err!=nil{
		// log.Fatal(err)
	     return nil,err
	}

	// defer db.Close() we will not close it until used 
	db.SetMaxOpenConns(1)
	//we will share one shared db connection 

	return db,nil
}

