
self.addEventListener('install', function(event) {
    var registration = self.registration
    var source = new EventSource('/sse?topics=log')
    source.addEventListener('log',function (event){
        console.log('Got log event from server')
        console.log('log event with message :'+event.data)
        const title = 'Server send event demo'
        const options = {
            body: `you have a message from server: ${event.data}`,
            tag:'sse-event',
            data:{
                url:'https://www.baidu.com',
            },
            renotify:true
        }
        // new Notification(title,options)
        registration.showNotification(title,options)
    })
})