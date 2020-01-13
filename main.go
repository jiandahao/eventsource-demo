package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/jiandahao/servemux"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

type (
	Message struct {
		Data string `json:"data"`
		Target string `json:"target"`
		Topic string `json:"topic"`
	}

	Consumer struct {
		conn net.Conn
		eMessage chan string
		topics string //*list.List
	}
)

var(
	topicConsumerList map[string]*list.List = make(map[string]*list.List )
)

func main(){
	//mesChannel = make(chan string)
	var router = servemux.NewRouter()

	router.UseStatic("/","./")
	router.Get("/sse",EventSourceHandler)
	router.Get("/sse2",EventSourceHandler2)
	router.Post("/send",SendNotification)
	if err := http.ListenAndServe(":8081",Log(router)); err != nil{
		log.Println("Error occurs when listening")
	}
}

func EventSourceHandler2(w http.ResponseWriter, r *http.Request){
	// Make sure that the writer supports flushing.
	//
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(":comment\n\n"))
	flusher.Flush()
	id := 0
	//isClosed := false
	msgChan := make(chan string)
	go func() {
		for{
			msgChan <- fmt.Sprintf("id:jian\nevent:log\ndata:hhaahahah %v\n\n",id)
			id++
			time.Sleep(5*time.Second)
		}
	}()
	for{
		select {
			case <- r.Context().Done():
				//isClosed = true
				fmt.Println("client done")
				fmt.Println(r.Context().Err())
				return

			case msg := <- msgChan:
				if _,err := w.Write([]byte(msg)); err!= nil{
					fmt.Printf("write content error client died: %s",err)
					return
				}
				fmt.Println(msg)
				flusher.Flush()
		}
	}
}

func newConsumer(w http.ResponseWriter, r *http.Request) (*Consumer, error){
	consumer := &Consumer{}
	conn  , _, err := w.(http.Hijacker).Hijack()
	if  err != nil{
		w.WriteHeader(502)
		w.Write([]byte("internal error"))
		return nil, err
	}

	// response with HTTP/1.1 200 OK, and with Content-Type : text/event-stream
	// more details refer to https://www.w3.org/TR/2009/WD-eventsource-20090421
	_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\n"))
	if err != nil{
		conn.Close()
		return nil, err
	}

	consumer.conn = conn
	consumer.topics = r.URL.Query().Get("topics")
	print(consumer.topics)
	if consumer.topics == ""{
		w.WriteHeader(400)
		w.Write([]byte("invalid argument, set the topic that you are interested"))
		return nil, fmt.Errorf("invalid argument")
	}
	consumer.eMessage  = make(chan string)
	// tell the client it's the end of message, otherwise client will pending until receiving an '\n\n'
	_, err = conn.Write([]byte("\n\n"))
	return consumer,nil
}


func Log(handler http.Handler) http.Handler{
	log.Println("Server is running at port :8081")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s %s\n",r.RemoteAddr,r.Method,r.Host,r.URL)
		handler.ServeHTTP(w, r)
	})
}
func EventSourceHandler(w http.ResponseWriter, r *http.Request){
	consumer , err := newConsumer(w,r)
	if err != nil{
		fmt.Println(err)
		return
	}
	if topicConsumerList[consumer.topics] == nil{
		topicConsumerList[consumer.topics] = &list.List{}
	}

	topicConsumerList[consumer.topics].PushBack(consumer)
	go func() {
		fmt.Println("start")
		for{
			select{
			case msg := <- consumer.eMessage:
				fmt.Println("Received a message")
				fmt.Printf(msg)
				consumer.conn.SetWriteDeadline(time.Now().Add(20*time.Millisecond))
				if _,err := consumer.conn.Write([]byte(msg));err != nil{
					fmt.Println("client closed with error :" + err.Error())
					return
				}
			}
		}
	}()
	go func() {
		// simulate message pushing
		id := 0
		for {
			id++
			consumer.eMessage <- fmt.Sprintf("id: %s\nevent:%s\ndata:%s\n\n",strconv.Itoa(id),"log","nihao")
			time.Sleep(10*time.Second)
		}

	}()
	fmt.Println("done")
}

var id int = 0
func SendNotification(w http.ResponseWriter, r *http.Request){
	message := Message{}
	if err := GetRequestBody(r,&message); err != nil{
		log.Println(err)
	}

	fmt.Println("sending topics")
	if message.Topic != "" && topicConsumerList[message.Topic] != nil{
		id++
		consumers := topicConsumerList[message.Topic]
		for consumer := consumers.Front(); consumer != nil; consumer = consumer.Next(){
			msg := fmt.Sprintf( "id: %s\nevent:%s\ndata:%s\n\n",strconv.Itoa(id),message.Topic,message.Data)
			fmt.Println(msg)
			c  := consumer
			go func() {
				c.Value.(*Consumer).eMessage <- msg
			}()
		}
	}

}

func GetRequestBody(request *http.Request, form interface{}) error{
	jsondata, _ := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	err := json.Unmarshal(jsondata, form)
	return err
}