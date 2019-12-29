package temp

import (
	"encoding/json"
	"github.com/jiandahao/servemux"
	"io/ioutil"
	"log"
	"net/http"
)
var(
	mesChannel map[string]chan string
	msg string
)

type (
	Message struct {
		Data string `json:"data"`
		Target string `json:"target"`
	}
)

func main(){
	//mesChannel = make(chan string)
	var router = servemux.NewRouter()

	router.UseStatic("/","./")
	router.Get("/sse",EventSourceHandler)
	router.Post("/send",SendNotification)
	if err := http.ListenAndServe(":8080",Log(router)); err != nil{
		log.Println("Error occurs when listening")
	}
}

func Log(handler http.Handler) http.Handler{
	log.Println("Server is running at port :8080")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s %s\n",r.RemoteAddr,r.Method,r.Host,r.URL)
		handler.ServeHTTP(w, r)
	})
}
func EventSourceHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","text/event-stream")
	w.Header().Set("Cache-Control","no-cache")
	w.Header().Set("Connection","keep-alive")

	//w.Write([]byte(":注释\n\n"))
	//w.Write([]byte("data:"+ "hahahahaha" + "\n\n"))
	//for{
	//	select  {
	//		case res := <- mesChannel :
	//			w.Write([]byte("data:"+ res + "\n\n"))
	//	}
	//}
	//if mesChannel == nil{
	//	mesChannel = make(map[string]chan string)
	//}
	if msg != ""{
		w.Write([]byte("data:"+msg + "\n"+"event:log"+"\n\n"))
		msg = ""
	}
	//if mesChannel["demo"] == nil{
	//	mesChannel["demo"] = make(chan string)
	//	res := <- mesChannel["demo"]
	//	fmt.Println("got")
	//	w.Write([]byte("data:"+res+"\n\n"))
	//}else{
	//	w.Write([]byte("waiting"))
	//}


}

func SendNotification(w http.ResponseWriter, r *http.Request){
	message := Message{}
	if err := GetRequestBody(r,&message); err != nil{
		log.Println(err)
	}
	if message.Target == ""{
		w.WriteHeader(400)
		return
	}
	//if mesChannel[message.Target] == nil{
	//	w.WriteHeader(401)
	//	return
	//}
	msg = message.Data
	log.Println("done")
	w.Write([]byte("success"))
}

func GetRequestBody(request *http.Request, form interface{}) error{
	jsondata, _ := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	err := json.Unmarshal(jsondata, form)
	return err
}