package drouter

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

type InitWrapper func(rw http.ResponseWriter, r *http.Request, i RequestInfo) (userdata interface{}, ok bool)
type Middleware func(userdataIn interface{}, rw http.ResponseWriter, r *http.Request, p RequestInfo) (userdataOut interface{}, ok bool)
type EndHandler func(userdata interface{}, rw http.ResponseWriter, r *http.Request, p RequestInfo)

type RequestInfo struct {
	Params httprouter.Params
	Route  string
}

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
	Verbose     bool
	InitWrapper *InitWrapper
	RouterNode
}

func (d *DRouter) InitRouter(router *httprouter.Router) error {

	var localInitHandler Middleware
	if d.InitWrapper != nil {
		localInitHandler = func(userdata interface{}, rw http.ResponseWriter, r *http.Request, p RequestInfo) (interface{}, bool) {
			var ok bool
			userdata, ok = (*d.InitWrapper)(rw, r, p)
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
		d.NextNodes[i].CreateRoutes(router, d.PathPart, &localInitHandler, d.Verbose)
	}

	return nil
}

// TODO release version without checks
//func (d *DRouter) InitRouterRelease() (*httprouter.Router, error)

func (n *RouterNode) CreateRoutes(router *httprouter.Router, path string, wrapper *Middleware, verbose bool) {
	var localwrapper Middleware = func(userdata interface{}, w http.ResponseWriter, r *http.Request, p RequestInfo) (interface{}, bool) {
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

	route := path + n.PathPart

	if n.EndPoint != nil {
		endpoint := func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
			var userdata interface{}

			info := RequestInfo{
				Params: p,
				Route:  route,
			}

			userdata, ok := localwrapper(userdata, rw, r, info)
			if !ok {
				return
			}
			n.EndPoint.Handler(userdata, rw, r, info)
		}
		router.Handle(n.EndPoint.Method, route, endpoint)
		if verbose {
			fmt.Printf("Route created: %s", route)
		}
	}

	for i := 0; i < len(n.NextNodes); i++ {
		n.NextNodes[i].CreateRoutes(router, route, &localwrapper, verbose)
	}
}
