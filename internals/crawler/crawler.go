package crawler


import (

	"fmt"
	"net/http"
	"net/html"
	"time"
	"io"
	"log"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)
func traverse(n *html.Node) {
    if n.Type == html.ElementNode && n.Data == "a" {
        for _, a := range n.Attr {
            if a.Key == "href" {
                fmt.Println(a.Val)
            }
        }
    }

    for c := n.FirstChild; c != nil; c = c.NextSibling {
        traverse(c)
    }
}

// func isValid(href):
//     if href starts with "#": return false
//     if href contains ":": return false
//     if not starts with "/wiki/": return false
//     return true
func Crawler(){
	client := &http.Client{
    Timeout: 5 * time.Second,
    }
	//basic building of scrper is this thing learn by doing

	req,err:=http.NewRequest("GET","https://en.wikipedia.org/wiki/Cristiano_Ronaldo",nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:149.0) Gecko/20100101 Firefox/149.0")
    

	resp,err:=client.Do(req)
	if err!=nil{
		panic(err)
	}

    contenttype:=resp.Header.Get("Content-Type")
	fmt.Println(contenttype)
    
	//log the first 10 lines of the content after getting the get request 
	defer resp.Body.Close()

	//prints the status of request where its going on to be 
	fmt.Println("Response Status",resp.Status)


	//  r := strings.NewReader("Hello World!")

	//  data,_:=io.ReadAll(r)

	// //now log the result of get request body
	// scanner:=bufio.NewScanner(data)
	//     for i := 0; scanner.Scan() && ; i++ {
    //     fmt.Println(scanner.Text())
    // }

	// if err := scanner.Err(); err != nil {
    //     panic(err)
    // }

	// Create a buffer for 500 bytes
	buffer := make([]byte, 5000)
	
	// Read up to 500 bytes from the body
	n, err := resp.Body.Read(buffer)
	
	// err is io.EOF if the body is smaller than 500 bytes,
	// but we still processed 'n' bytes.
	if err != nil && err != io.EOF {
		panic(err)
	}

	// Print the read characters
    bodyToread:=string(buffer[:n])
	fmt.Println(bodyToread)
	//its the callers responsibilty that we provide 
	//a UTF-8 encoded hTML FILE STRUCTRE 
	//Creating a tokenizer and a parser
	//for us n is that 
	// z:=html.NewTokenizer(r)
	// for{
	// 	tt:=z.Next()
	// 	if tt==html.ErrorToken{
	// 		panic("something wrong in tokenizer")
	// 	}

	// 	//processing here 
	// 	doc,err:=html.parse(z)
	// 	if err != nil {
	// 	log.Fatal(err)
	// }
	// for n := range doc.Descendants() {
	// 	if n.Type == html.ElementNode && n.DataAtom == atom.A {
	// 		for _, a := range n.Attr {
	// 			if a.Key == "href" {
	// 				fmt.Println(a.Val)
	// 				break
	// 			}
	// 		}
	// 	}
	// }

	//we will not use a tokenizer based approach 
	docs,err :=html.Parse(n) //io.rEDAER
	if err!=nil{
		log.Fatal(err)
	}
   
	traverse(docs)


	



}