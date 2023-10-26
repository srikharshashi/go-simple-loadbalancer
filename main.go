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

// simpleServer implements server
type simpleServer struct{
	address string
	proxy *httputil.ReverseProxy
}

func  (s *simpleServer) Address() string {
	return s.address
}

func (s *simpleServer) IsAlive() bool{
	return true
}

func(s *simpleServer) Serve(w http.ResponseWriter,r *http.Request){
	s.proxy.ServeHTTP(w,r)
}



type LoadBalancer struct {
	port string
	roundRobinCount int
	servers []Server
}

func newLoadBalancer(port string,severs []Server) *LoadBalancer {
	return &LoadBalancer{
		port: port,
		servers: severs,
		roundRobinCount: 0,
	}
}

//methods on loadbalancer
func (loadBalancer *LoadBalancer) getNextAvialableSever() Server{
	// pick a sever here
	// if the server is not alive then go on picking other severs until you find a sever that is alive
	// then return that sever
	server := loadBalancer.servers[loadBalancer.roundRobinCount%len(loadBalancer.servers)]
	for !server.IsAlive(){	
		loadBalancer.roundRobinCount++;
		 server = loadBalancer.servers[(loadBalancer.roundRobinCount)%len(loadBalancer.servers)]
	}
	loadBalancer.roundRobinCount++;
	return server
}

func (loadBalancer *LoadBalancer)  serveProxy(rw http.ResponseWriter,r *http.Request){
	targetServer := loadBalancer.getNextAvialableSever()
	fmt.Printf("forwarding requests to address %q\n",targetServer.Address())
	targetServer.Serve(rw,r)

}

func newsimpleServer(address string) *simpleServer{
	severUrl,err:=url.Parse(address)
	handleErr(err)

	return &simpleServer{
		address: address,
		proxy: httputil.NewSingleHostReverseProxy(severUrl),
	}

}

func handleErr(err error){
	if(err!=nil){
		fmt.Printf("Error: %v\n ",err)
		os.Exit(1);
	}
}

func main(){
	servers:=[]Server{
		newsimpleServer("https://www.facebook.com"),
		newsimpleServer("https://www.bing.com"),
		newsimpleServer("https://www.duckduckgo.com"),

	}

	lb:= newLoadBalancer("8000",servers)
	handleRedirect := func(w http.ResponseWriter,r *http.Request){
		lb.serveProxy(w,r)
	}  
	http.HandleFunc("/",handleRedirect)

	fmt.Printf("serving requests at localhost:%s\n",lb.port)
	http.ListenAndServe(":"+lb.port,nil)

}