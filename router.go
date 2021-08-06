package drouter

import (
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

type InitWrapper func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) (userdata interface{}, ok bool)
type Middleware func(userdataIn interface{}, rw http.ResponseWriter, r *http.Request, p httprouter.Params) (userdataOut interface{}, ok bool)
type EndHandler func(userdata interface{}, rw http.ResponseWriter, r *http.Request, p httprouter.Params)

type RouterNode struct {
	PathPart  string
	EndPoint  *EndPoint
	Wrapper   *Middleware
	NextNodes []RouterNode
}

type EndPoint struct {
	Method  string
	Handler EndHandler
}

type DRouter struct {
	Host        string
	NextNodes   []RouterNode
	InitPath    string
	InitHandler *InitWrapper
}

func (d *DRouter) InitRouter() (*httprouter.Router, error) {
	router := httprouter.New()

	var localInitHandler Middleware
	if d.InitHandler != nil {
		localInitHandler = func(userdata interface{}, rw http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, bool) {
			var ok bool
			userdata, ok = (*d.InitHandler)(rw, r, p)
			if !ok {
				return userdata, false
			}
			if reflect.ValueOf(userdata).Kind() != reflect.Ptr {
				panic("userdata must be a pointer")
			}
			return userdata, true
		}
	}
	for i := 0; i < len(d.NextNodes); i++ {
		d.NextNodes[i].CreateRoutes(router, d.InitPath, &localInitHandler)
	}

	return router, nil
}

//func (d *DRouter) InitRouterRelease() (*httprouter.Router, error)

func (n *RouterNode) CreateRoutes(router *httprouter.Router, path string, wrapper *Middleware) {
	var localwrapper Middleware = func(userdata interface{}, w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, bool) {
		var ok bool
		if wrapper != nil {
			userdata, ok = (*wrapper)(userdata, w, r, p)
			if !ok {
				return nil, false
			}
		}

		if n.Wrapper != nil {
			userdata, ok = (*n.Wrapper)(userdata, w, r, p)
			if !ok {
				return nil, false
			}
		}

		return userdata, true
	}

	if n.EndPoint != nil {
		endpoint := func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
			var userdata interface{}
			userdata, ok := localwrapper(userdata, rw, r, p)
			if !ok {
				return
			}
			n.EndPoint.Handler(userdata, rw, r, p)
		}
		router.Handle(n.EndPoint.Method, path+n.PathPart, endpoint)
	}

	for i := 0; i < len(n.NextNodes); i++ {
		n.NextNodes[i].CreateRoutes(router, path+n.PathPart, &localwrapper)
	}
}
