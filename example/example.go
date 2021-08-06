package main

import (
	"net/http"

	"github.com/royalcat/drouter"
)

type RequestData struct {
	Token string
}

func main() {
	var requestInit drouter.InitWrapper = func(rw http.ResponseWriter, r *http.Request, i drouter.RequestInfo) (userdata interface{}, ok bool) {
		rw.Header().Set("Content-Type", "application/json")
		println("init handler")
		return &RequestData{
			Token: "dada",
		}, true
	}

	var tokenWrapper drouter.Middleware = func(userdata interface{}, rw http.ResponseWriter, r *http.Request, i drouter.RequestInfo) (interface{}, bool) {
		//tokenStr := r.Header.Get("Authorization")[7:]
		println("token wrapper")
		requestData := userdata.(*RequestData)
		requestData.Token = "aaa"

		return requestData, true
	}

	var typesHandler drouter.EndHandler = func(userdata interface{}, rw http.ResponseWriter, r *http.Request, i drouter.RequestInfo) {
		println("endpoint")
		rw.Write([]byte("AAAAA"))

	}

	routes := drouter.DRouter{
		Host:        "localhost",
		InitPath:    "/v1",
		InitHandler: &requestInit,
		NextNodes: []drouter.RouterNode{
			{
				Wrapper: &tokenWrapper,
				NextNodes: []drouter.RouterNode{
					{
						PathPart:  "/user/:userId",
						NextNodes: []drouter.RouterNode{},
					},
					{
						PathPart: "/cars",
						EndPoint: &drouter.EndPoint{
							Method:  "GET",
							Handler: typesHandler,
						},
					},
					{
						PathPart: "/sensors",
					},
				},
			},
		},
	}

	router, _ := routes.InitRouter()

	http.ListenAndServe("localhost:8988", router)

}
