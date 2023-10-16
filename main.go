package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)
type Server interface{
Address() string
IsAlive() bool
Serve(w http.ResponseWriter,r *http.Request)
}
type SimpleServer struct{
addr string
proxy *httputil.ReverseProxy
}
type LoadBalancer struct{
port string
roundRobinCount int
servers []Server
}
func newSimpleServer(addr string) *SimpleServer{
serverUrl,err:=url.Parse(addr)
handleError(err)
return &SimpleServer{
addr: addr,
proxy: httputil.NewSingleHostReverseProxy(serverUrl),
}
}
func handleError(e error){
if e!=nil{
fmt.Printf("Error : %v\n",e)
os.Exit(1)
}
}
func NewLoadBalancer(port string,servers []Server) *LoadBalancer{
return &LoadBalancer{
port:port,
roundRobinCount: 0,
servers: servers,
}
}
func (s *SimpleServer) Address() string{
return s.addr
}
func (s *SimpleServer) IsAlive() bool{
return true
}
func (s *SimpleServer) Serve(w http.ResponseWriter,r *http.Request){
 s.proxy.ServeHTTP(w,r)
}
func (lb *LoadBalancer) getNextAvailableServer() Server{
 server:=lb.servers[lb.roundRobinCount%len(lb.servers)]
 for !server.IsAlive(){
lb.roundRobinCount+=1
server=lb.servers[lb.roundRobinCount%len(lb.servers)]
}
lb.roundRobinCount+=1;
return server
}
func (lb *LoadBalancer) serveProxy(w http.ResponseWriter,r *http.Request){
targetServer:=lb.getNextAvailableServer()
fmt.Printf("forwarding request to address %s\n",targetServer.Address())
targetServer.Serve(w,r)
}
func main(){
servers:=[]Server{
newSimpleServer("https://www.facebook.com"),
newSimpleServer("http://www.bing.com"),
newSimpleServer("http://www.duckduckgo.com"),
}
lb:=NewLoadBalancer("8000",servers)
handleRedirect:=func(w http.ResponseWriter,r *http.Request){
lb.serveProxy(w,r)
}
http.HandleFunc("/",handleRedirect)
fmt.Printf("server is serving at 'localhost :%s'\n",lb.port)
http.ListenAndServe(":"+lb.port,nil)
}