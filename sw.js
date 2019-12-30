self.addEventListener('message',function (event) {
    console.log('Got message from server')
})

self.addEventListener('log',function (event) {
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
    new Notification(title,options)
})