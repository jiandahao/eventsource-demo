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
	router.Post("/send",SendNotification)
	if err := http.ListenAndServe(":8081",Log(router)); err != nil{
		log.Println("Error occurs when listening")
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
	_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\n"))
	if err != nil{
		conn.Close()
		return nil, err
	}
	_, err = conn.Write([]byte("\r\n"))
	_, err = conn.Write([]byte("\r\n"))
	_, err = conn.Write([]byte("\r\n"))
	consumer.conn = conn
	consumer.topics = r.URL.Query().Get("topics")
	print(consumer.topics)
	if consumer.topics == ""{
		w.WriteHeader(400)
		w.Write([]byte("invalid argument, set the topic that you are interested"))
		return nil, fmt.Errorf("invalid argument")
	}
	consumer.eMessage  = make(chan string)
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
	//isReady := make(chan bool)
	go func() {
		//<-isReady
		fmt.Println("start")
		for{
			select{
			case msg := <- consumer.eMessage:
				fmt.Println("Received a message")
				consumer.conn.Write([]byte(msg))
			}
		}
	}()
	//go func() {
	//	for {
	//		consumer.eMessage <- fmt.Sprintf("id: %s\n event:%s\ndata:%s\n\n","1","log","nihao")
	//		time.Sleep(2*time.Second)
	//	}
	//
	//}()
	//consumer.conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\n"))
	//w.Write([]byte(":comment"))
	fmt.Println("done")
	//isReady <- true
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
			print(consumer)
			msg := fmt.Sprintf("id: %s\n event:%s\ndata:%s\n\n",strconv.Itoa(id),message.Topic,message.Data)
			//fmt.Println(msg)
			c  := consumer
			go func() {
				print(msg)
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